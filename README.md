# pkiplot - Turn your cert-manager PKI into Mermaid Diagrams

pkiplot (pronounced _pikaplot_) is a small utility that reads [cert-manager](https://cert-manager.io/)
`Certificates`, `Issuers` and `ClusterIssuers` from YAML files and outputs a Mermaid diagram that describes
the entire PKI.

## Example: kcp

[kcp](https://www.kcp.io/)'s Helm chart sets up a less than trivial PKI in its Helm chart. If you render the
Helm chart and then convert it with pkiplot:

```
cd kcp-dev/helm-charts/charts
helm template --namespace kcp kcp ./kcp | pkiplot -n kcp -
```

… you receive this output:

```mermaid
graph TB
    i_kcp/kcp_client_issuer([kcp-client-issuer]):::issuer
    i_kcp/kcp_etcd_client_issuer([kcp-etcd-client-issuer]):::issuer
    i_kcp/kcp_etcd_peer_issuer([kcp-etcd-peer-issuer]):::issuer
    i_kcp/kcp_front_proxy_client_issuer([kcp-front-proxy-client-issuer]):::issuer
    i_kcp/kcp_pki([kcp-pki]):::issuer
    i_kcp/kcp_pki_bootstrap([kcp-pki-bootstrap]):::issuer
    i_kcp/kcp_requestheader_client_issuer([kcp-requestheader-client-issuer]):::issuer
    i_kcp/kcp_server_issuer([kcp-server-issuer]):::issuer
    i_kcp/kcp_service_account_issuer([kcp-service-account-issuer]):::issuer
    c_kcp/kcp(kcp):::cert
    c_kcp/kcp_ca(kcp-ca):::ca
    c_kcp/kcp_client_ca(kcp-client-ca):::ca
    c_kcp/kcp_etcd(kcp-etcd):::cert
    c_kcp/kcp_etcd_client(kcp-etcd-client):::cert
    c_kcp/kcp_etcd_client_ca(kcp-etcd-client-ca):::ca
    c_kcp/kcp_etcd_peer(kcp-etcd-peer):::cert
    c_kcp/kcp_etcd_peer_ca(kcp-etcd-peer-ca):::ca
    c_kcp/kcp_external_admin_kubeconfig(kcp-external-admin-kubeconfig):::cert
    c_kcp/kcp_front_proxy(kcp-front-proxy):::cert
    c_kcp/kcp_front_proxy_client_ca(kcp-front-proxy-client-ca):::ca
    c_kcp/kcp_front_proxy_kubeconfig(kcp-front-proxy-kubeconfig):::cert
    c_kcp/kcp_front_proxy_requestheader(kcp-front-proxy-requestheader):::cert
    c_kcp/kcp_front_proxy_vw_client(kcp-front-proxy-vw-client):::cert
    c_kcp/kcp_internal_admin_kubeconfig(kcp-internal-admin-kubeconfig):::cert
    c_kcp/kcp_pki_ca(kcp-pki-ca):::ca
    c_kcp/kcp_requestheader_client_ca(kcp-requestheader-client-ca):::ca
    c_kcp/kcp_service_account(kcp-service-account):::cert
    c_kcp/kcp_service_account_ca(kcp-service-account-ca):::ca
    c_kcp/kcp_virtual_workspaces(kcp-virtual-workspaces):::cert

    i_kcp/kcp_etcd_peer_issuer --> c_kcp/kcp_etcd_peer
    i_kcp/kcp_front_proxy_client_issuer --> c_kcp/kcp_external_admin_kubeconfig
    i_kcp/kcp_client_issuer --- c_kcp/kcp_front_proxy_kubeconfig --> c_kcp/kcp_internal_admin_kubeconfig
    i_kcp/kcp_requestheader_client_issuer --- c_kcp/kcp_front_proxy_requestheader --> c_kcp/kcp_front_proxy_vw_client
    i_kcp/kcp_service_account_issuer --> c_kcp/kcp_service_account
    i_kcp/kcp_server_issuer --- c_kcp/kcp --- c_kcp/kcp_front_proxy --> c_kcp/kcp_virtual_workspaces
    i_kcp/kcp_etcd_client_issuer --- c_kcp/kcp_etcd --> c_kcp/kcp_etcd_client
    i_kcp/kcp_pki --> c_kcp/kcp_ca
    i_kcp/kcp_pki --> c_kcp/kcp_client_ca
    i_kcp/kcp_pki --> c_kcp/kcp_etcd_client_ca
    i_kcp/kcp_pki --> c_kcp/kcp_etcd_peer_ca
    i_kcp/kcp_pki --> c_kcp/kcp_front_proxy_client_ca
    i_kcp/kcp_pki_bootstrap --> c_kcp/kcp_pki_ca
    i_kcp/kcp_pki --> c_kcp/kcp_requestheader_client_ca
    i_kcp/kcp_pki --> c_kcp/kcp_service_account_ca
    c_kcp/kcp_client_ca --> i_kcp/kcp_client_issuer
    c_kcp/kcp_etcd_client_ca --> i_kcp/kcp_etcd_client_issuer
    c_kcp/kcp_etcd_peer_ca --> i_kcp/kcp_etcd_peer_issuer
    c_kcp/kcp_front_proxy_client_ca --> i_kcp/kcp_front_proxy_client_issuer
    c_kcp/kcp_pki_ca --> i_kcp/kcp_pki
    c_kcp/kcp_requestheader_client_ca --> i_kcp/kcp_requestheader_client_issuer
    c_kcp/kcp_ca --> i_kcp/kcp_server_issuer
    c_kcp/kcp_service_account_ca --> i_kcp/kcp_service_account_issuer

    classDef clusterissuer color:#7F7
    classDef issuer color:#77F
    classDef ca color:#F77
    classDef cert color:orange
```

## Installation

Either [download the latest release](https://github.com/xrstf/pkiplot/releases) or build for yourself using Go 1.20+:

```bash
go install go.xrstf.de/pkiplot
```

## Usage

Couldn't really be any simpler:

```bash
Usage of pkiplot:
      --cluster-resource-namespace string   cert-manager's cluster resource namespace, used to find secrets referenced by cluster-scoped objects (default "cert-manager")
  -f, --format string                       Output format (one of [mermaid]) (default "mermaid")
  -n, --namespace string                    Only include namespace-scoped resources in this namespace (also the default namespace for resources without namespace set)
  -v, --verbose                             Enable more verbose output
  -V, --version                             Show version info and exit immediately
```

## License

MIT
