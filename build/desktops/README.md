# Building kVDI Desktops

This directory contains references/exampls for creating "Desktop" images for `kVDI`.

The examples in this directory are not intended as a "you must do it this way", but moreso just some patterns that can be followed.

## Desktop Requirements

At the end of the day only one thing needs to be running to allow `kVDI` to create a session with containers booted from your image.
A `VNC` server listening on a `tcp` or `unix` socket.

The examples in this directory use `tigervnc` as the VNC server for a couple of reasons:

 - Can create its own `DISPLAY` using a dummy X. No need to run a separate X server.
 - Out-of-the-box remote resolution resizing

Once the VNC server is running, you can really run anything on the `DISPLAY` that you'd like.

## Process Control System

The two examples in this directory so far each take a different approach to the process control.

### Ubuntu/Tini/SupervisorD

The images based off `ubuntu` use `supervisord` (behind `tini`) as the "init" system for the desktops.

The configuration is relatively straight forward.
The base images install and configure X utilities and the VNC server, and the extended images just add the extra packages and processes they want.

One of the main benefits of this approach is no requirement for extra kernel capabilities.
However, the configuration files may not be as straight-forward as some, and the desktop environments end up as direct descendants of the root process in the image.

#### _UPDATE_

In the middle of porting these images to systemd. So they are largely defunt and updates are not being pushed at the moment. The published ones still work and can be extended.

### Arch/Systemd

The `archlinux` based images use `systemd` as an init system inside the container.
This gives you quite a bit of flexibility with having almost the entire `systemd` ecosystem at your disposal.

The configuration is a bit more complex, but at the end of the day you are still just defining processes to run at launch.
Some `pam`/`getty` hackery in the base image will cause a `systemd --user` process to spawn when the container launches, and then essentially any process enabled with `systemctl enable --global --user my-process.service` will start with the containr, and running as the desktop user.

This does provide more process isolation to a degree, but at the expense of requiring `CAP_SYS_ADMIN` on the pods themselves.

#### _UPDATE_

Having `systemd` as a user-process control system available made the audio support much easier to accomplish.

Will be looking into using it for all images, but at the time of writing have been having trouble getting it to work in ubuntu.


These images are a little harder to test.
The `kVDI` manager will take care of all needed configurations at launch time,
but for running locally you will need to ensure the proper mounts and temp filesystems that `systemd` expects.
