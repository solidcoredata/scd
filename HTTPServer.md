# HTTP Server

SCD will eventually be made into units that can be integrated into other
systems and configuration tools. However for now start with a simple design.

## scdhttp

scdhttp
  -config "" "directory to discover configuration files"

Types of configs:

 * LoadUnit
  - HTTP
  - Auth
  - App1
  - UI Widgets
  - Stats
  - Logs
 * Pipeline
  - HTTP bind then
  - Auth then
  - App1 handle

The scdhttp can register various units. It itself can register and use
an HTTP server, but functionally it is the same as any other unit that
is loaded.

In the future some of these methods of loading would be replaced with k8s
and service discovery.

The goal for this iteration is to have a few executables that provide
unique functionality to the app as a whole. The app itself will be defined
with a directory of configs.

Later additions will be offline tools for schema and config validation
and schema commits.


## units

### scdload

When a unit exec loads up, if the "-scdunit" flag is given then it will write
to standard out a JSON string. The config would include the port it is listening
on and various configs that it supports.

```go
type StdOutConfig struct {
	Bind string // "localhost:8402"
	Units []struct {
		Kind string // "SearchListDetail"
		Schema ...
	}
}
```


```json
{
	Bind: "localhost:9999",
	Units: [
		{
			Kind: "scdload",
			Schema: ??,
		},
		{
			Kind: "scdhttp",
			Schema: ??,
		},
		{
			Kind: "scdpipe",
			Schema: ??,
		},
	],
}

{
	Kind: "scdhttp",
	Name: "mysitehttp",
	Binds: [
		{On: "mysite.com:https"},
	],
}
```

### scdauth

```json
{
	Bind: "server1:24245",
	Units: [
		{
			Kind: "scdauth",
			Schema: ??,
		}
	],
}
{
	Kind: "scdauth",
	Next: [
		{On: "Authorized", Send: "appA1"},
		{On: "ChangePassword", Send: "updatePass"},
		{On: "Un-Authorized", Send: "login"},
		{On: "NeedU2F", Send: "u2f"},
	],
}
```

### scdui

```json
{
	Bind: "server1:24246",
	Units: [
		{
			Kind: "searchlistdetail",
			Schema: ??,
		},
		{
			Kind: "menuscreen",
			Schema: ??,
		}
	],
}

{
	Kind: "searchlistdetail",
	Name: "accountDef",
	Source: "table:Account",
	Fields: [
		{Type: "Group"},
		{Type: "Text", Field: "Name", Display: "Name"},
	],
	Detail: {
		Actions: [
			{Display: "Run Proc1", Run: "Proc1"},
		]
	}
}
```

### appA1

```json
{
	Bind: "server1:24247",
	Units: [
		{
			Kind: "appA1",
			Schema: ??,
		}
	],
	Intrinsic: [
		{
			Kind: "proc",
			ProcName: "Proc1",
		}
	]
}

{
	Kind: "menuscreen"
	Helpers: [
		"scddownload",
		"scdnotify",
		"scdupdate",
	],
	MenuProperties: {
		Location: "top",
	}
	Menu: [
		{
			Name: "Admin",
			Menu: [
				{Name: "Accounts", Run: "screen:accountDef"},
			],
		},
	],
}
```


## Network / API process

HTTP receives request.
	can:
	 * request info to attach to context
	 * route based on rules / code
	 * 
