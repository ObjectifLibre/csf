# How to write CSF reactions

## JSON reaction format

```json
{
    "name": "reaction_name",
    "event": "event_name",
    "script": "console.log('this script will run before any action and decides what to do based on event data'); if (event['key'] == 42)  {nextAction = 'first_action'}  result = event;",
    "actions": 	{
	"first_action": {
	    "script": "console.log('this script will run *after* the action exec and decides what to exec next'); if (actionData['key'] == 'value') {nextAction = 'second_action'}",
	    "action": "dummy_action",
	    "module": "dummy"
	},
	"second_action": {
	    "script": "",
	    "action": "dummy_action",
	    "module": "dummy"
	}
    }
}

```

A reaction is composed of:

 - `name` is the name of the reaction (be creative !)
 - `event` is the name of the event the reaction respond to (a GET request to `http://$CSF_ADDR:8888/v1/events` will give you everything you need)
 - `script` is the JS script that will run first before any action. Use it to decide what to do based on the event data
 - `actions` is the different actions that can be executed or not

An action is composed of:

 - `module`: the name of the module containing the action
 - `action`: the name of the action (a GET to `http://$CSF_ADDR:8888/v1/actions` will help you)
 - `script`: script that choose to continue the pipeline or not. Executed *after* the action.

## Scripting

Some variables you'll need:

 - `event` is the data sent by the event
 - `nextAction`: set this to choose which action to run next, leave empty to stop the pipeline
 - `err`: set this to any non-empty string and the pipeline will stop with your error message
 - `result`: set this to pass data to an action
 - `actionData`: the output data of the action

Every variable can be used in any script exec except `actionData` which is not available in the first `script` exec.