# Yarn Berry Test App

This is a test application for the Yarn Berry buildpack support. It uses:

- Yarn Berry (v4.0.0) as specified in `packageManager` field
- `.yarnrc.yml` configuration file  
- Berry-format `yarn.lock` file
- PnP (Plug'n'Play) node linker

The app serves a simple HTTP server that uses the `leftpad` dependency.
