# Webmesh Desktop

A Virtual Desktop Infrastructure on top of Webmesh.

**It is fast because it is built on top of Webmesh, which is powered by Wiregaurd™️**

**It is scalable for all needs because it runs on Kubernetes**

**It is easy to use!**


## Features

  - Containerized user desktops running on Kubernetes with no virtualization required (`libvirt` options may come in the future). All traffic between the end user and the "desktop" is encrypted.
  - Persistent user data
  - Audio playback and microphone support
  - File transfer to/from "desktop" sessions. Directories get archived into a gzipped tarball prior to download.
  - RBAC system for managing user access to templates, roles, users, namespaces, serviceaccounts, etc.
  - MFA Support
  - Configurable backend for internal secrets. Currently `vault` or Kubernetes Secrets
  - Use built-in local authentication, LDAP, or OpenID.
  - App metrics to either scrape externally or view in the UI. More details in the `helm` doc.