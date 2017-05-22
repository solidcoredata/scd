# Solid Core Data Design

Solid Core Data is a framework designed to remove some of the
repetative tasks associated with building a new SQL backed line
of business application. This framework is not intended
to directly dictate or improve UI or SQL. This framework is 
intended to provide default ways to authenticate and authorize
users, move data between the backend and the frontend, and
a way to create reusable UI components through typed and composable
UI  configurations.

## Motivation

Line of business applications can be constructed from base UI
components, a database server, and a HTTP server. If that is done
many things end up being re-invented for each application created.
This is true even when larger frameworks are used such as
Angular or Hibernate. These type of frameworks work great when
a custom model is defined for each screen, but by using a custom
model for each screen reusability drops. Also many UI components
available are either not designed for data by either not being
visually dense enough or they do not perform will with large
data sets.

In short, Solid Core Data aims to:

 * Choose or create UI components that work will with large
   data sets.
 * Use data marshalling techniques that can be effectivly
   reused.
 * Implement an authentication and authorization component
   that can be used out of the box or replaced with a
   well defined interface.
 * Implement a way to handle SQL queries.

## Design and Configuration

### Frontend

The frontend is a combination of well defined components
and application specific configuration that configures
and composes components together. For instance, the frontend
framework will provide a search list detail component that
itself is composed of two forms and a grid component. The
grid and form configurations specify one or more fields
to bind to and what type of field component to use.

An application can still create custom screen types and
components. A component can be created using most any
UI platform. The only criteria is that it satisfies
a perticular interface and takes a configuration
when creating it.

### Backend

The backend is made up of:

 * HTTP server
 * Authentication and authorization server
 * UI server
 * Query server
 * Report server
 * Application server

Each one of these are defined through network friendly
interfaces. While default implementations will be provided
they can also be replaced with alternate implementations.

### Random details

 * Data is sent from the backend to the frontend using a
   standard table set based format.
 * Changes are encoded as row deltas.

