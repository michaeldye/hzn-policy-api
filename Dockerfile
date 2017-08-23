FROM busybox:1.27-glibc

EXPOSE 8091

ADD hzn-policy-api /

VOLUME /policy.d/
VOLUME /config/

WORKDIR /config/

CMD [ "/hzn-policy-api", "-v", "5", "-logtostderr", "-configfile", "config.toml" ]
