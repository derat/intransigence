# apparmor

This directory contains example [AppArmor] configuration files for restricting
what the `intransigence` executable can do.

The files can be copied to their corresponding locations under
`/etc/apparmor.d`. After customizing files for local paths, load the profile by
running `apparmor_parser -r /etc/apparmor.d/intransigence`.

[AppArmor]: https://wiki.ubuntu.com/AppArmor
