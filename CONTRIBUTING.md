# Contributing

## Prerequisites

This project uses [Mise](https://mise.jdx.dev/) for managing development tools and dependencies.

## Setup

1. Clone the repository
2. Run `mise setup` to install required tools (Go, hk, pkl)
3. Create a `.env` file for API credentials (see `.env.example` for reference)

## Development

Build the project:

    mise build

Run tests against the live API:

    mise test

Run the CLI:

    mise rom-tools -- [args...]
