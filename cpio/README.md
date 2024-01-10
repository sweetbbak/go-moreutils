# Cpio
Create cpio archives for use in an initramfs

```sh
find . > namelist
cpio -ov < namelist > output.cpio
```

This version unfortunately does NOT support reading input from a pipe...
ie:
```sh
# DOES NOT WORK
find . | cpio -ov > output.cpio
```
