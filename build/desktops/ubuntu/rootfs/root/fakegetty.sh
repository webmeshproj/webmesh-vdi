#!/bin/bash

0<&- script -qefc "/sbin/agetty --autologin ${USER} --noclear - xterm"
