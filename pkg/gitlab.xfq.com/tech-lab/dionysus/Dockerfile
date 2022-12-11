FROM reg.weipaitang.com/micro/base:alpine.3.8

ADD dionysus /dionysus
ADD conf /conf
ADD activate/engine_test.yaml /conf/config.yaml

WORKDIR /
ENTRYPOINT [ "/dionysus", "gin" ]