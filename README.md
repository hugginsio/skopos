# skopos

I have a simple problem: I run my own DNS distributed across a handful of Kubernetes clusters. They are exposed to my Tailnet via the Tailscale Operator and set as the primary
resolvers within the Tailnet. However, sometimes there are availability issues: the Tailnet IP might change, an authentication key might expire, or a site might experience network
unavailability. In order to maintain performant connectivity, I want unhealthy resolvers to be automatically removed from the Tailnet configuration.
