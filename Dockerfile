FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY config/ /idea-go/config/
ADD ./http-server /idea-go/http-server
RUN  chmod u+x /idea-go/http-server
WORKDIR /idea-go
EXPOSE 50051
COPY run-http.sh /run-http.sh
RUN chmod u+x /run-http.sh
CMD ["/run-http.sh"]