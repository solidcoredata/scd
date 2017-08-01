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
   - [ ] Client router. Use `history.pushState`, state/page should be encoded in query segment of URL.
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