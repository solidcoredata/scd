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
			Schema: {
				Types: [
					{
						Name: "Menu",
						Members: [
							{Type: "Text", Name: "Name"},
							{Type: "[]Menu", Name: "Menu", Nullable: true}
							{Type: "Text", Name: "Run", Nullable: true}
						],
					},
					{
						Name: "LocationEnum",
						Values: ["Top", "Left"],
					}
				],
				Root: [
					{Name: "Menu", Type: "Menu"},
					{Name: "MenuProperties", Type: [
						{Name: "Location", Type: "LocationEnum"},
					]},
				],
			},
		},
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

### Client Request Life-cycle:

 * GET / RETURN HTML with common loader script.
 * Script sees URL State and requests that state from the server.
 * Server returns a list component configurations. Each configuration includes the component URL (which serves as the component unique ID).
   If the component is not found it may be loaded from the given URL.
 * The client applies all of the components sent back along with the configurations requested.
 * It would be ideal to be able to "nest" components logically, but still maintain a "context" bucket and allow for deep linking.
 * A component will manage its own internal state (values set, etc) but still need to expose them to allow them to be settable for deep links.
   - I'm not sure this is possible or reasonable.

### HTTP Request Life-cycle:

 * Client sends request to server.
 * HTTP Server receives request.
 * HTTP Server takes credential token(s) and send the token(s) to the Authentication Server to establish authentication, roles, and login state.
   The result is attached to the request context.
   - Login state (Logged Out, U2F Login Wait, Must Change Password, Logged In)
   - Elevated state (Normal, Elevated Login)
 * HTTP Server sends entire request with context to application back-end, switching off of URL and Login State.
 * ...
 
### Thinking

 * Not all changes to the UI should go through the server.
 * Components may opt to provide either a linked component or a linked call that will provide a component.
 * It is the responsibility of the parent component to attach a linked component to the DOM / Context Data.
 * Thus the initial loader is actually just a really basic component that bootstraps other components.

 * There is a link between UI hierarchy, UI Layout, Data Context, and deep linking state.

Application Endpoints

 * /api/proc
 * /api/ui
 * /api/delta
 * /api/query
 * /api/lookup
 * /api/error # Errors encountered while on the UI.
 * /api/login
 * /api/logoff

The Delta should be able to send a Validate request that the server can check in real-time.

### Components that build on each other

  * Router
    - Login API, static assets.
      - Skinnable default presentation. Basic UI, settable name and logo. Optional Reset Password link.
    - Query, Delta, Lookup, Proc APIs. Static assets. Component loader API. Logout and Elevate APIs.
      - Default application loader to bootstrap core components.
	  - Application frame / sreen menu.
	    - Components such as Grid and Form Groups.
	    - Skinnable compositions.

Sane defaults for business applications. But applications can still replace stacks as required.

No inheritance, only configuration and systems that use other systems.

