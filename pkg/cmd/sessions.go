/*

   Copyright 2020,2021 Avi Zimmerman

   This file is part of kvdi.

   kvdi is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   kvdi is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with kvdi.  If not, see <https://www.gnu.org/licenses/>.

*/

package cmd

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tinyzimmer/kvdi/pkg/api/client"
	"github.com/tinyzimmer/kvdi/pkg/types"
)

var (
	createSessionOpts types.CreateSessionRequest
	proxyHost         string
	proxyPort         int
)

func init() {
	createFlags := sessionCreateCommand.Flags()
	createFlags.StringVar(&createSessionOpts.Template, "template", "", "the template to launch")
	createFlags.StringVar(&createSessionOpts.Namespace, "namespace", "", "the namespace to launch the template in")
	createFlags.StringVar(&createSessionOpts.ServiceAccount, "service-account", "", "a service account to attach to the session")

	sessionCreateCommand.MarkFlagRequired("template")
	sessionCreateCommand.RegisterFlagCompletionFunc("template", completeTemplates)
	sessionCreateCommand.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		nss, err := kvdiClient.GetNamespaces()
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}
		return nss, cobra.ShellCompDirectiveDefault
	})
	sessionCreateCommand.RegisterFlagCompletionFunc("service-account", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ns := createSessionOpts.Namespace
		if ns == "" {
			ns = "default"
		}
		sas, err := kvdiClient.GetServiceAccounts(ns)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveError
		}
		return sas, cobra.ShellCompDirectiveDefault
	})

	proxyFlags := sessionsProxyCmd.PersistentFlags()
	proxyFlags.StringVar(&proxyHost, "host", "127.0.0.1", "the host to bind the listener to")
	proxyFlags.IntVar(&proxyPort, "port", 5900, "the port to bind the listener to")

	sessionsProxyCmd.AddCommand(sessionDisplayProxyCmd)
	sessionsProxyCmd.AddCommand(sessionAudioProxyCmd)

	sessionsCmd.AddCommand(sessionsGetCmd)
	sessionsCmd.AddCommand(sessionCreateCommand)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsProxyCmd)
	sessionsCmd.AddCommand(sessionCopyCmd)
	sessionsCmd.AddCommand(sessionStatCmd)

	rootCmd.AddCommand(sessionsCmd)
}

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"sess", "session", "s"},
	Short:   "Desktop sessions commands",
}

var sessionsGetCmd = &cobra.Command{
	Use:               "get [SESSION]",
	Short:             "Retrieve VDI sessions",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeSessions,
	PreRunE:           checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := kvdiClient.GetDesktopSessions()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			for _, sess := range sessions.Sessions {
				if sess.NamespacedName() == args[0] {
					return writeObject(sess)
				}
			}
			return fmt.Errorf("No session %q found", args[0])
		}
		return writeObject(sessions.Sessions)
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:               "delete [SESSIONS...]",
	Short:             "Terminate VDI sessions",
	Aliases:           []string{"del", "rem", "rm", "remove", "term", "terminate"},
	ValidArgsFunction: completeSessions,
	PreRunE:           checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			nn, err := argToNamespacedName(arg)
			if err != nil {
				return err
			}
			if err := kvdiClient.DeleteDesktopSession(nn); err != nil {
				return err
			}
			fmt.Printf("Session %q terminated\n", nn.String())
		}
		return nil
	},
}

var sessionCreateCommand = &cobra.Command{
	Use:     "create",
	Short:   "Launch a VDI session",
	Aliases: []string{"new"},
	PreRunE: checkClientInitErr,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := kvdiClient.CreateDesktopSession(&createSessionOpts)
		if err != nil {
			return err
		}
		return writeObject(resp)
	},
}

var sessionsProxyCmd = &cobra.Command{
	Use:     "proxy",
	Aliases: []string{"serve"},
	Short:   "Proxy VDI sessions to the local machine",
}

var sessionDisplayProxyCmd = &cobra.Command{
	Use:               "display",
	Aliases:           []string{"video", "vid"},
	Short:             "Proxy a session's display",
	PreRunE:           checkClientInitErr,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSessions,
	RunE: func(cmd *cobra.Command, args []string) error {
		nn, err := argToNamespacedName(args[0])
		if err != nil {
			return err
		}
		fmt.Println("Retrieving display connection to", nn.String(), "(if this takes a while a connection may already be open)")
		conn, err := kvdiClient.GetDesktopDisplayProxy(nn)
		if err != nil {
			return err
		}
		return proxyConn(conn)
	},
}

var sessionAudioProxyCmd = &cobra.Command{
	Use:               "audio",
	Aliases:           []string{"sound", "snd"},
	Short:             "Proxy a session's audio",
	PreRunE:           checkClientInitErr,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSessions,
	RunE: func(cmd *cobra.Command, args []string) error {
		nn, err := argToNamespacedName(args[0])
		if err != nil {
			return err
		}
		fmt.Println("Retrieving audio connection to", nn.String(), "(if this takes a while a connection may already be open)")
		conn, err := kvdiClient.GetDesktopAudioProxy(nn)
		if err != nil {
			return err
		}
		return proxyConn(conn)
	},
}

func proxyConn(conn io.ReadWriteCloser) error {
	defer conn.Close()
	addr := net.JoinHostPort(proxyHost, strconv.Itoa(proxyPort))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening for connections on", addr)
	for {
		clientConn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting client connection:", err.Error())
		}
		fmt.Println("-- Handling connection from", clientConn.RemoteAddr().String())
		go io.Copy(clientConn, conn)
		io.Copy(conn, clientConn)
		clientConn.Close()
		fmt.Println("Client", clientConn.RemoteAddr().String(), "disconnected --")
	}
}

var sessionStatCmd = &cobra.Command{
	Use:     "stat",
	Aliases: []string{"ls"},
	Short:   "List files and directories in a VDI session",
	PreRunE: checkClientInitErr,
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		sessions, err := kvdiClient.GetDesktopSessions()
		if err != nil || len(sessions.Sessions) == 0 {
			return []string{}, cobra.ShellCompDirectiveError
		}
		out := make([]string, len(sessions.Sessions))
		for i, sess := range sessions.Sessions {
			out[i] = fmt.Sprintf("%s/%s:", sess.Namespace, sess.Name)
		}
		if len(strings.Split(toComplete, ":")) < 2 {
			return out, cobra.ShellCompDirectiveNoFileComp
		}
		for _, opt := range out {
			if !strings.HasPrefix(opt, toComplete) {
				continue
			}
			return completeSessionPath(toComplete)
		}
		return out, cobra.ShellCompDirectiveNoSpace
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		spl := strings.Split(args[0], ":")
		if len(spl) < 2 {
			return fmt.Errorf("%q is not a valid session path", args[0])
		}
		sess := spl[0]
		path := strings.Join(spl[1:], ":")
		nn, err := argToNamespacedName(sess)
		if err != nil {
			return err
		}
		res, err := kvdiClient.StatDesktopFile(nn, path)
		if err != nil {
			return err
		}
		return writeObject(res)
	},
}

var sessionCopyCmd = &cobra.Command{
	Use:     "copy",
	Aliases: []string{"cp", "scp"},
	Short:   "Copy files to and from a VDI session",
}

func argToNamespacedName(arg string) (client.NamespacedName, error) {
	spl := strings.Split(arg, "/")
	if len(spl) != 2 {
		return client.NamespacedName{}, fmt.Errorf("%q is not a valid session name", arg)
	}
	return client.NamespacedName{
		Namespace: spl[0],
		Name:      spl[1],
	}, nil
}
