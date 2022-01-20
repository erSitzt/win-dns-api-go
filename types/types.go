package types

type DnsZone struct {
	ZoneName       		string
	ZoneType       		string
	IsAutoCreated  		string
	IsDsIntegrated 		string
	IsReverseLookupZone string
	IsSigned 			string
}

type DnsRecord struct {
	Name  string
	Type  string
	TTL   int
	Value string
	Aging int
}
