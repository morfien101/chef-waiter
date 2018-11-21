# This file contains fixes that can't go anywhere else.
# Fixes in here need to be runable repeatably with out error.

# This fix is used to update the Upstart script for chefwaiter.
# An issue was identified where the service was not starting on CentOS 6
# servers. The was due to either the service starting to soon or Upstart
# 0.6.5 not function correctly as documented.
# The rc service is SysV which means that chefwaiter will start
# as one of the last services.
sed -i 's/start on filesystem or runlevel \[2345\]/start on stopped rc/' vendor/github.com/kardianos/service/service_upstart_linux.go