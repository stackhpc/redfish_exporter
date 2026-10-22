FROM python

RUN pip install sushy-tools

ADD sushy.crt .
ADD sushy.key .
ADD mockups/ mockups/

EXPOSE 443/tcp

# This CMD is overriden during tests
CMD ["sushy-static", "-p", "443", "-c", "sushy.crt", "-k", "sushy.key", "-m", "mockups/DSP2043-server"]
