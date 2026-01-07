# High-Level Architecture

You will build two coordinated services:

1. DNS Monitor (primary application)
2. Service Monitor (Kubernetes-resident component)

 The DNS Monitor will be responsible for monitoring the tailnet DNS configuration and performing health checks. On container boot, it will pull the tailnet DNS configuration into an internal state tracker and then continuously perform healthchecks to sync the internal state with the tailnet DNS configuration. Whenever a healthcheck fails, the resolver is removed from internal state which is then synchronized back to the tailnet DNS configuration. The healthcheck mechanism does not actually check if the resolver is working, merely if the resolver node is still connected to Tailscale.

- [ ] dnsmon connects to Tailscale
- [ ] resolver inventory is initialized from current tailnet state
- [ ] healthcheck daemon which runs periodically
- [ ] resolver inventory synchronizes back to tailnet on change
- [ ] RPC server allows for manual triggering of healthcheck, inventory synchronization, and node addition/removal
- [ ] inventory synchronization includes getting all nodes in the tailnet tagged as DNS resolvers and adding them

While this handles resolver removal, to facilitate the addition of new resolvers we will deploy a service monitor inside Kubernetes clusters alongside the DNS resolvers. The service monitor will watch for the creation and removal of cluster-local Service resources that bear the "tailscale.com/tags=tag:dns" annotation. Whenever such a Service is created, `svcmon` will make an RPC call to `dnsmon` that adds the node to the resolver inventory and thus the tailnet DNS configuration. Similarly, when such a Service is removed, `svcmon` will make an RPC call to `dnsmon` to remove the node from the resolver inventory and tailnet DNS configuration.

- [ ] svcmon connects to Tailscale
- [ ] svcmon connects to Kubernetes
- [ ] svcmon watches for Service creation and removal in the configured namespace with the configured annotation.
- [ ] svcmon communicates service creation to dnsmon
- [ ] svcmon communicates service removal to dnsmon
- [ ] svcmon lists existing Services on startup and communicates them to dnsmon

With this architecture, the tailnet DNS configuration should be eventually-consistent with active Kubernetes DNS resolvers. If an entire cluster goes down, the DNS configuration will rectify itself on the next healthcheck cycle. If connectivity is restored without the `svcmon` restarting, `dnsmon` will find the restored resolvers upon the next inventory reconciliation loop.
