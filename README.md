# Solid Core Data Design

Solid Core Data is a framework designed to remove some of the
repetitive tasks associated with building a new SQL backed line
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
 * Use data marshalling techniques that can be effectively
   reused.
 * Implement an authentication and authorization component
   that can be used out of the box or replaced with a
   well defined interface.
 * Implement a way to handle SQL queries.

## High Level Design

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
a particular interface and takes a configuration
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

### Development

Applications are comprised of server-side code, schema definitions,
schema transition plans, UI definitions, and UI custom components.

Applications are compiled from a set of application services and definitions
to a deployable unit. The primary target (and only target at this time) is
the HTTP server. The compilation step verifies all queries are valid,
and all fields correclty reference valid columns.

### Non-goals

 * Don't try to make development faster. This might happen,
   but only as a side effect.
 * Don't try to change how development is done. Use git, standard SQL,
   allow the use standard UI libraries and components.
 * Don't try to control everything. Assume each application will have
   some custom controls and screens. Assume application may need
   to provide different backend servers or wrap an existing authentication
   or reporting server.

## Project Roadmap

The project will aim for three phases:

 1. Build MVP (v0.5)
 2. Build out optional components (v0.10)
 3. Use framework in several projects and adjust and tweak (v0.11 to v1.0)

Before the project is used in a real business project, create an official business
for solid core data and assign copyright to it with appropriate CLAs.
Put up a website at solidcoredata.com for the business and solidcoredata.org for
the project. The business will at least initially be to separate out personal
interests from project interests and provide businesses using the framework
an official point of contact.

Present in a MVP release:

 * HTTP server
 * UI server
  - With a performant search list detail component and fields.
 * Query server
 * Example application server
 * Simple authentication and authorization server

Building out optional components will include:

 * Make a configuration editor
 * Report server
 * More function in authentication and authorization server
 * SQL to SQL Query server
 * More UI Components

Using the framework and various tweaks is TBD.

## Implementation

### Security

To report a security bug plase email [kardianos@gmail.com](mailto:kardianos@gmail.com).
Please set the subject `SCD Security`.
Your email will be acknoledged within 72 hours.

### License

Source code files should not list authors names directly.
Each file should have a standard header:
```
// Copyright 2017 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
```

At the moment copyright is not assigned. However before external contributions
are accepted or the framework used in production, a business must be formed
and copyright assigned directly to the project. The Solid Core Data project will
be made distinct from the business name.

