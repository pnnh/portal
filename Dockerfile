
FROM golang:1.24 AS base

FROM base AS deps

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

FROM deps AS builder
WORKDIR /app
COPY . .

# Build Portal
RUN go build -o ./portal

FROM base AS runner
WORKDIR /app

RUN addgroup --system --gid 1001 golang
RUN adduser --system --uid 1001 portal --ingroup golang

RUN chown -R portal:golang /app
COPY --from=builder --chown=portal:golang /app/portal .

USER portal

ENV PORT=8001
ENV HOSTNAME="0.0.0.0"

CMD ["./portal", "-config", "env://CONFIG"]
