---
title: Backend And Architecture
---

# Backend And Architecture

## Backend Systems

At Nirvana Labs, Bruno has worked on backend systems for a modern cloud platform serving AI, gaming, blockchain, and real-time workloads.

He designed and shipped Go API resources such as Projects, Organizations, Memberships, and role-based access patterns for multi-tenant customer accounts. He also worked on workflows with Temporal for creating resources and coordinating backend operations.

A major project was the migration from Auth0 to WorkOS across the dashboard and API. This involved authentication, identity, enterprise SSO, SCIM directory sync, and SIEM audit logging.

At Notabene, Bruno helped build and improve backend systems for Travel Rule compliance, transaction workflows, address risk scoring, sanctions screening, third-party integrations, and Kafka-based services.

At Banco do Brasil, Bruno worked on Open Finance and API modernization, helping expose banking services through secure REST APIs.

## APIs, Microservices, Queues, And Event-Driven Systems

Bruno has strong experience with APIs, especially REST APIs, secure APIs, third-party integrations, and internal backend services.

At Nirvana Labs, he works with Go APIs, GORM, multi-tenant resources, identity flows, and authentication and authorization patterns.

At Notabene, he worked with APIs, microservices, Kafka, background workers, third-party provider integrations, transaction workflows, and customer-facing compliance features.

At Banco do Brasil, he worked on banking APIs, OAuth 2.0 security patterns, and Open Finance developer platforms.

## Reliability, Performance, And Scalability Improvements

At Notabene, Bruno improved how large CSV exports were handled. Instead of processing large exports synchronously through the API, the work was moved to an asynchronous worker flow using Kafka. That made the system more reliable and avoided blocking API requests with heavy processing.

He also helped evolve parts of the Notabene system from a monorepo toward more focused services with Kafka. This improved service boundaries and helped the system scale better as the product grew.

At Nirvana Labs, Bruno has worked on identity and resource-creation flows where reliability matters because they affect customer access, organizations, projects, memberships, and enterprise authentication.

## Database Design

Bruno thinks database design should start from the product model and access patterns.

He usually asks what the core entities are, how they relate to each other, how they will be queried, what needs to be consistent, and what may become difficult to migrate later.

He likes relational databases for most product systems because they make relationships, constraints, and business rules easier to reason about. He cares about clear schemas, good indexes, explicit migrations, and keeping the data model understandable.

At Nirvana Labs, resources like organizations, projects, memberships, and roles need to be modeled clearly because they affect access control and multi-tenant behavior.
