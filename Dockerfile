FROM scratch
MAINTAINER Brian Hechinger <wonko@4amlunch.net>

ADD perms-srv-linux-amd64 perms-srv
VOLUME /etc/chremoas

ENTRYPOINT ["/perms-srv", "--configuration_file", "/etc/chremoas/chremoas.yaml"]
