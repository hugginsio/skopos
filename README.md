# skopos

I have a simple problem: I run my own DNS distributed across a handful of Kubernetes clusters. They are exposed to my Tailnet via the Tailscale Operator and set as the primary resolvers within the Tailnet. However, sometimes there are availability issues. In order to maintain performant connectivity, I want unhealthy resolvers to be automatically removed from the Tailnet configuration.

Skopos solves this problem by providing two monitoring components:

1. `dnsmon`, which manages the Tailnet DNS configuration based on health checks
2. `svcmon`, which runs alongside my resolvers and manage DNS configuration based on cluster state
