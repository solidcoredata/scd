# General Design

## Run v Compile

Developers should not write applications that get directly run.

 1. Correctness is a primary goal. Verification needs a separate step to run.
 2. Database fields and UI fields need to be checked at a compile time
    to ensure they are correct and consistent. Queries should be checked.
 3. Database fields will sometimes (often) drive other field lists.
 4. A database schema can have more information attached to the schema
    then just data types and names. Field groupings, accessability,
	and useage patterns may be part of the schema that is used
	by the final application.

At some point in the future it should be possible to create think client
controls and replicate the web application on a mobile or desktop device.
This would involve altering the compiler output in some way. The HTTP server
would not be used to send the application, it would just be used for
API access (though it would probably be different then the HTTP server API).

### Compile

An application is built up in layers:

 * Database Schema
 * Queries
 * (UI widgets) Not part of the specific app stack exactly, but required by
     the UI definition. Should be able to validate the UI definition.
 * UI Definition. Must know what database coloumns and query columns it uses
     so compiler can create correct alter script and to detect errors early.

### Run (HTTP)

There is a well defined router pattern, data requester, and data format.
The server widget instances are provided to the server which are
sent to the client on demand. All the base components for every SPA
is loaded from the server.

## Next

 * [ ] Create the client components of a SPA:
   - [ ] Client router. Use URL Hash section. Do not use history API, enables links to just be links.
   - [ ] Encode client state outside of a widget.
   - [ ] Data cache.
   - [ ] Basic navigation frame.
   - [ ] Basic widget.
   - [ ] Nested or Modal wdget.
 * [~] Make an API that can register resources to serve.
   - Each resource should state if it is a single resource or tree, if staic or dynamic.
 * [~] Add Error login state, used for mis-configured or system errors to display.
 * [x] Split out the UI and Data portions, allow composing AppHandlers into various
       parts. Create a registration process.
 * [ ] Allow AppHandlers to request resources from other handlers.
 * [ ] Determine real-time message API too, something that on the backend is a
   GRPC streaming response, and on the front end is a HTTP POST long pull.
   This will be used early on for updating the client automatically during
   development. Later it may be used to indicate a report is completed.
   Avoid websockets.
 * Components won't create API or deal with requests and responses I think.
   - Components need to work with some internally defined APIs.
   - Components need to hit their own backend database.
 * Look into Kubernetes API aggregation layer / Custom Resource Definition.
   SCD needs to allow applications to extend and have custom data points.
   This may have specific ideas on how to design "extension" within SCD.
   - https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
   - https://kubernetes.io/docs/concepts/api-extension/apiserver-aggregation/

## Mixed Notes

 * In new API, want to be able to lazy load values before edit, optionally show
   a preview for gird.
   - Used for lazy load PDF or binary preview.
   - Used for editing long / large string values.
 * Data is sent from the backend to the frontend using a
   standard table set based format.
 * Changes are encoded as row deltas.
 * Backend servers should be (micro-)service friendly
   and easily hostable on k8s.
 * Authentication server will support FIDO U2F out of the box.
 * When developing an application, the application needs to satisfy
   one or more interfaces to be run under the framework. The framework
   itself can be used as a binary blob and never touched or compiled
   during application development.
 * Dev frame (around normal nav frame) can be used to update other components.
 * Updates can be pushed by server to client if dev mode is on.
 * Application compiles to a servable spec that contains database schemas,
   alter scripts to increment schema versions, UI definitions, UI widgets,
   and custom code.
 * Each version of the system needs to be checkpointed and saved off for
   updates, including the UI and widget set.

## Router Notes

The authenticator needs to be per URL, NOT per RouteHandler or app.
On second thought, the authentication system should be tied to a single cookie name.
The reason I wanted per URL cookie names is to prevent programs running on different ports
from overwriting each other's cookies. One way to prevent this and to work around this
at least for now is to use a consistent hash scheme of the Hostname.

The RouteHandler should not have Authenticator, it should be on the LoginStateRouter.
Additionally the LoginStateRouter should be able to pass along the authenticator to
all applications.

I'd like to be able to support multiple QA environments on the same Host.
I can probably do this by configuring a special LoginGranted handler that
maps paths to nested-applications.

In addition to the State handler, I'd also like to define various service
descriptions that can be implemented by a service. Then we define a
list of services that implement service descriptions.
We then move the Authenticator into a service description
and move the session manager into a service description. The
AuthenticateMemory service is used to satisfy both service descriptions.
The API component handlers for login and logout both require the service
description for session management.

Service include: Authenticator, Session Manager, Database Querier,
Report Engine, Scheduler, Job Runner.

The service registry could itself be a component, able to be swapped out.
The service registry could be a name mapper.

Another component that should be first class is the notifier.

Need a (central?) configuration manager for all components.
Need to allow a well known configuration be the basis
for other varients with extra properties.
Need to register a type of configuration, then delcare configuration
instances.
Configuration of the current running instance should be separate from
the new running instance. Need to be able to transition away from
one system to another gracefully with GOAWAY and similar messages.

It would be good to have a standardized way to handle file resources:
templates, images, HTML, favicon, and any other client or server side
resources (anothing not code).

Widget Registry. Component A loads widget instances 1,2,3 into the registry.
Component B loads widget instances 4,5,6 into the registry.
Data sources need to be able to be routed to the correct database.
A Widget needs to have a handle to both the database and the table.
When a system that contains widgets are loaded, we also record the
databases used by that system. So we actually have "service + database + table"
to record for data routing.

	1. Create Thing1 that has a database handle + schema.
	2. Create a registry for Thing1.
	3. Let Thing1 pipe a ui widget to the client.
	4. Let /api/data send data requests to Thing1's database.

 * Route Handler / HTTP Server:
    Configured on HTTP Server. Next step a caddy plugin probably.
	Each server needs an HTTP listener, an authenticator service, and one or more application services.
	gRPC is exlusivly used beyond the HTTP server.
 * Authenticator Service:
	Referenced by multiple services.
	Provides login / logout / session auth services via API.
	In future may also provide an application screen for managing users and setup.
 * Application Services:
   - URL Sharding:
		Composing URL routes is easy, will rely on getting notified when
		The end service restarts or updates URLs.
		URL conflicts should not bring down everything, just the routes in question.
   - SPA Widgets & Configurations:
		Have a specialized SPA registry.
		Probably each application will register one or more prefixs to forward to
		the application, as well as specific names (esp for boot strap names).
   - Data Handlers:
		Probably similar in many ways to the above.
		Details TBD.

Route Handler / HTTP Server holds the configuration to the authentication
service and application services. It also configures the login type, prefix,
and consume-redirect directive.

Once connected application services will need to notify the Route Handler
of the sharding rules, route assignments. Notifications on updates will be important.

TBD: how to handle k8s style rolling update, or what the story is on
versioning application servers.

---

What are the different possible ways to configure / connect various components together?

 * Put all configuration in the router.
   - Upside: this simplifies how it is built and maintained.
   - Downside: that isn't really what I want. I'd like to be able to plug a new
     application into a router and have it just run.
 * Have the HTTP Router be really dumb, call out to the Login State Router.
   - How to configure the Login State Router? I just moved the problem around without
     gaining much.
 * I really want to reference some type of "bundle" (that can reference other bundles)
   at a server location.
   - A bundle would be a collection of "Login State -> Component".
   - A bundle may also optionally register a new Login State like "Change Password" or "Login".
   - Some components need to test for capabilities like "supports U2F", delcare this somehow.
   - The authenticator would be set at the Login router level and be injected into
     any component that requires it (as a service endpoint).
   - Every component needs to have some unique string identifier for registration and routing.

From the last point:

HTTP Router Input:

 * HTTP hostname, cert, other HTTP related configuration.
 * Athenticator service name for each Login Router.
 * A list of services and bundle names.
   - service-1 fetch bundle-A and bundle-B
   - service-2 fetch bundle-Z

Router -> Applications (bundle-A, bundle-B, bundle-Z) returns
```
Configure LoginState
	LoginState: None,prefix=/login/,consume-redirect=false
	LoginState: Granted,prefix=/app/,consume-redirect=true
Configure Components
	LoginState=None: service-1/NoneSessionAPI
		URL=/api/login
	LoginState=None: service-2/LoginUI
		URL=/
		URL=/lib/
	LoginState=Granted: service-1/GrantedSessionAPI
		URL=/api/logout
	LoginState=Granted: service-1/SPAAPI
		URL=/api/fetch-ui
		URL=/api/fetch-data
	LoginState=Granted: service-1/SPAUI
		URL=/
		URL=/base/lib/
		SPA-Code=base
		SPA-Code=navigate
		SPA-Code=widget-1
		SPA-Code=widget-1
	LoginState=Granted: service-2/component
		SPA-Code=custom-widget-9
		SPA-Config=xyz
```

Ensure no conflicts arise. If they do, disable both conflicting components,
but still load the application as much as possible.

Each service so far in examples have been called "service-1" or "service-2".
But what is a service identifier? Each service has a well known name, such as:

 * solidcoredata.org/library-1
 * solidcoredata.org/example-1
   - references: solidcoredata.org/library-1

The Router needs to resolve the well known names into addresses. How it does so
will depend on the environment.

An entire application will end up specified with a single bundle. Then the Router
specification would look like:
```
bind=example-1.solidcoredata.org:https
bundles=[solidcoredata.org/example-1/app]
resolve=[
	solidcoredata.org/example-1 = localhost:8002,
	solidcoredata.org/library-1 = localhost:8001,
]
```

Next: Example of nested bundles. Show how Login State may sometimes be stated
      and sometimes not be stated in a bundle.

---

Current design calls for a component to serve up the SPA code and SPA configurations.
Ideal to keep this, but unsure how to do so without exesive hops.

 1. service-2 gets a request for widget-1
 2. service-2 requests widget-1 from Router
 3. Router requests widget-1 from service-1
 4. Router sends widget-1 to service-2
 5. service-2 returns widget-1 in original request

or...

 1. After a configuration event, router hands list of services to service-2
 2. service-2 gets a request for widget-1
 3. service-2 requests widget-1 from service-1
 4. service-1 returns widget-1 to service-2
 5. service-2 returns widget-1 to original request

Yes, the second option is what we want to do.
It is likely that we can have the Router be a generic registry that updates
dependents on configuration events.
The endpoints need to know the types (gRPC intefaces), but the Router does not.

---

A service may need to respond differently to the same request values if the request
comes from different applications. For example Auth server needs to know if
the application is the PROD, QA, or USER:<user-1> environment, as each have different
authorized parties. Query servers will also need to know what the environment and
application it is from.

It seems like when the bundles are being created, configurations need to be
attached. When bundle items are created the configuration items that are expected
should be declared. Configuration values should not be application names; configuration
values should be:

 * environment (PROD, QA, USER:<user-1>)
 * data source (pg://db-server:3000/db1, sqlserver://db-server:4000/db2)
 * maybe widget configuration instances too

To define some terms that leak useage:

 * Potential Resource: a resource that may not be directly used to compose
   an application. It must first be paired with a configuration to be used.
   Examples include a widget type, an authentication service, a query service,
   a database.
 * Configured Resource: a resource that may be used to compose an application.
   Examples include a widget to list specific numbers, a QA authenticator, a
   specific database instnace with credentials.
 * Potential Bundle: a collection of Potential Resources. It may not include
   any Configured Resource. All the Ptential Resources must contain the same
   type of configuration.
   An example is an authenticator, a logout URL, and a login URL.
 * [Configured] Bundle: a collection of Configured Resources. It may not include
   any Potential Bundles or Potential Resources.
   Eventually an application serves a single configured bundle.

All items are given unique names. A Potential resource "search-list-detail" widget
and a specific configuration is used to create the new configured resource
"user-list". The "user-list" can now be referenced directly.


