#!/bin/sh
# Crush chroot environment script with analytics disabled
export CRUSH_DISABLE_METRICS=1
export DO_NOT_TRACK=1

# Example usage for chroot environment
# chroot /some/path ./crush "$@"
crush "$@"