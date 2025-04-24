FROM debian:12

RUN apt update &&\
apt install -y socat

RUN apt-get update

RUN useradd --home-dir /home/ret2popacola --create-home ret2popacola
RUN mkdir /home/ret2popacola/chall

COPY ./ret2popacola /home/ret2popacola/chall/
COPY flag.txt /home/ret2popacola/chall/
RUN chmod 555 /home/ret2popacola/chall/ret2popacola

WORKDIR /home/ret2popacola/chall
USER ret2popacola

EXPOSE 4444

CMD socat tcp-listen:4444,reuseaddr,fork exec:/home/ret2popacola/chall/ret2popacola
