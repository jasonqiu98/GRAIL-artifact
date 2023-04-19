FROM arangodb:3.9.10

COPY --from=golang:1.19.4-alpine /usr/local/go/ /usr/local/go/
 
ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk update && apk add build-base