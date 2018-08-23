# Write your own module

## Event source

### Interface

Event sources send events to CSF via a simple go channel. An event is represented by the EventData struct:

```go
type EventData struct {
   Name string
   Data map[string]interface{}
}
```

You'll need to implement the `EventSourceInterface`:

```go
type EventSourceInterface interface {
   Setup(ch chan EventData, config []byte) error
   Events() map[string][]ArgType
}
```

Events() returns a list of a hard-coded map of type `map[string][]eventsources.ArgType` describing each event.
Setup() is called once at startup if your module is enabled in the configuration. Each module should store the event channel.

### Quick start

Simply copy `dummy_source/dummy_source.go`, replace ervery `dummy` with a good name for your module, edit the `Events()` function accordingly and replace randomDummyEventGenerator by your own function that will feed events to csf.


## Action module

### Interface

Action modules provide one or more actions to csf. An action generally takes input data, do something and can return data if needed. Input and output data is passed as a generic map of type `map[string]interface{}`. Your action module needs to implement the `ActionModuleInterface` interface:

```go
type ActionModuleInterface interface {
   Setup(config []byte) error
   Actions() (map[string][]ArgType, map[string][]ArgType)
   ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error)
   }
```
Setup() is called once at startup if your module is enabled in the configuration.
Actions() returns a list of hard-coded maps, one to describe input data and the other to describe output data.
ActionsHandler() returns the function that will be called by CSF whenever an action needs to be executed.

### Quick start

Copy `dummy_action/dummy_action.go`, replace every `dummy` occurance with a good name for your module, edit the `Actions()` function accordingly and replace `dummyActionHandler` with your own action handler.