FROM python:3.11-slim

RUN mkdir /tmp/userdbservice
COPY . /tmp/userdbservice/
WORKDIR /tmp/userdbservice/

RUN pip3 install -r requirements.txt
RUN pip3 install cryptography

RUN useradd -r admin -u 1000 \
    && echo 'admin:password124' | chpasswd

EXPOSE 5000
ENTRYPOINT [ "./deploy.sh" ]