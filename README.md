# Windows DNS API (GoLang)
===========

This is a simple API based on the [Win DNS API (Node. JS)](https://github.com/vmadman/win-dns-api). It has been rewritten in GoLang mainly for learning purpose and because it seemed more appropriate for this scenario.

This tool acts as an API for Windows Server DNS. With this it is possible to create/edit/delete DNS entries on a Windows Server.
To run this as a service take a look at [NSSM](http://nssm.cc/)

Compiled versions available for Windows 32 and 64 Bits on the [releases section](https://github.com/marcotuna/win-dns-api-go/releases)

This README will be updated as the project grows. Any contributions are welcomed!

# Basic usage

The API defines these URLs:

* `/`

    A simple welcome page.

* `/dns/`

    Lists all DNS zones known to this DNS server.

* `/dns/<zoneName>`

    Lists all DNS records in a given Zone.

* `/dns/<zoneName>/<dnsType>/<nodeName>/set/<ipAddress>`

    Creates or updates a DNS record. Note this will replace _all_ existing records for the given `nodeName` and `type`.

* `/dns/<zoneName>/<dnsType>/<nodeName>/remove`

    Removes all records of the given `type` for the given `nodeName`.

# Running Win-DNS-API-Go for testing

Compile using `GOOS=windows go build`, and just run the `.exe`:

```
C:\Users\your-dns-admin> win-dns-api-go.exe
Listening on port 3111.
```

# Running win-dns-api-go as a Windows service

The binary can also be installed as a Windows service using this command in a `cmd` session started as Administrator.
Note the space behind the equal signs, these are important:

```
sc create win-dns-api-go start= auto binPath= C:\Users\your-dns-admin\win-dns-api-go.exe
```

This will create a service named `win-dns-api-go` that you can use to run this tool in the background.

If you want to delete the service again, run:

```
sc delete win-dns-api-go
```
