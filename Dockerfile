FROM alpine
WORKDIR /opt/middleman
COPY middleman /opt/middleman/
COPY middleman.yml /etc/middleman/middleman.yml
EXPOSE 9723
CMD ["/opt/middleman/middleman", "--cfg.path=/etc/middleman/middleman.yml"]
