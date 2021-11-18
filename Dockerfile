FROM golang:1.17.3 AS build-stage

RUN mkdir /app
WORKDIR /app
COPY . .
RUN ./scripts/release.sh

FROM scratch AS export-stage
COPY --from=build-stage /app/dist/ /