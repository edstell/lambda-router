# lambda-router
Lightweight wrapper for routing lambda events to handling functions.

## Usage
Initialize a router and register handlers, then pass to lamda.Start.
```
router := New()
router.Route("Do", handleDo)
lambda.Start(router.Handle)
```

### Custom errors
If your error values contain rich content which you want to make available to
calling lambda functions, provide your own error marshaling function.
```
router := New(MarshalErrorsWith(func(error) ([]byte, error) {
	// Marshal your error.
	return marshaled, nil
}))
```
