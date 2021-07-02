package types

type DnsZone struct {
	Name       string
	Type       string
	Storage    string
	Properties []string
}

type DnsRecord struct {
	Name  string
	Type  string
	TTL   int
	Value string
	Aging int
}
