# Deployments

## QA

Use nested applications to auto-deploy all branches to a QA application.
A special application type uses sub-applications to represent each
application QA branch. The main application is a chooser for the desired branch.

## Migrating Persistent Data Schemas / Breaking Changes

Breaking changes are never allowed in a single deployment.
For instance, if a column that is currently being used in an application should
be removed, there are two steps: 1) remove all references to it in the UI and
backend 2) remove the column. The server is automated to first upgrade to the
first version, ensure all clients have it, then upgrade to the second version.

Various aspects can be determined through some type of upgrade policy service.
Determine who is on what version, and when it is safe to update through.
Determine max time policies for upgrades.

Changes that would break the application if deployed must fail their verification.

## Integrating Applications

This is a motivational story about two applications. App A is deployed in a
private intranet. App B is deployed to the public cloud for consumers to access.
As time goes on the buisiness wants to expose one or two specific functions from
(private) App A to (public) App B.

To accomplish exposing functions from App A into App B the developer does the
following:

 1. Establish a private gRPC proxy that allows App B to connect to
    App A's internal service.
 2. Configure App A to use App B's component that should be exposed.
 3. If custom skinning is needed, an encapsulating component may be
    designed for App B that uses App A's compontent service.

