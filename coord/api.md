# Coordination Server API

This API provides the spec for a given device and whether the device has applied the latest spec.

## Methods

All methods require an `Authorization` header with the following format:
```
Authorization: QrystalCoordIdentityToken <token>
```

### Get Latest Status

Method: Get
Path: `/v1/reify/{network}/{device}/latest`
Response: `application/json`, JSON of type `spec.GetReifyLatestResponse`

Returns whether the device has applied the latest spec. This should be polled regularly to receive updates.

### Get Spec

Method: Get
Path: `/v1/reify/{network}/{device}/spec`
Response: `application/json`, JSON of type `spec.NetworkCensored`

### Patch Spec

Method: Patch
Path: `/v1/reify/{network}/{device}/spec`
Request Body: `application/json`, JSON of type `coord.PatchReifySpecRequest`
Response: nothing

### Post Spec Application

Method: Post
Path: `/v1/reify/{network}/{device}/status`
Request Body: `application/json`, JSON of type `coord.PostReifyStatusRequest`
Response: `application/json`, JSON of type `coord.PostReifyStatusResponse`

Returns whether the applied spec is up-to-date.
