FROM ghcr.io/jmelfi/stargazer:latest

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
