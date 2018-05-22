FROM alpine
WORKDIR /opt/middleman
COPY middleman_linux_amd64 /opt/middleman/
COPY middleman.yml /etc/middleman/middleman.yml
EXPOSE 9723
CMD ["/opt/middleman/middleman_linux_amd64", "--cfg.path=/etc/middleman/middleman.yml"]
