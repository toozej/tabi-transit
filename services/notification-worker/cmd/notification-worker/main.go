// notification-worker deliberately starts disabled until D-017 push provider
// evidence and D-018 notification policy are approved. Runtime composition of
// PostgreSQL, key material, and an Expo PushGateway belongs to that enablement
// change; this binary never falls back to a real network sender.
package main

func main() {}
