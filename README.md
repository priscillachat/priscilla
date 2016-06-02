# Priscilla

Priscilla is a chat bot written in go. Most other popular chat bots today are
written in interpreted language. In my opinion, it's a bit unpractical to write
a chat bot in go the same way because it'll requires source code modification
and re-compilation every time you want to add functionalities to the bot. So I
took a bit different approach.

## Server-Adapter-Responder Model

Priscilla is made of three components: **Priscilla server**, **Priscilla
Adapter**, and **Priscilla Responder**.

### Priscilla Server

In reality, Priscilla server is actually just a message dispatcher server that
acts as the courier to facilite the communication between adapters (connect to
chat services) and responders (processes messages, performs actions, and
respond to messages). Communication protocol is in JSON over TCP and is
described later in the document. Though mostly stable, it's still in early
development stage and may change without notice until code base is stable.

### Priscilla Adapter

Adapters are long running Priscilla clients that connects and listens on the chat
services then forward messages to Priscilla server, and listen for messages from
Priscilla server and forward them back to the chat service.

An unique feature about Priscilla, is that since Priscilla does not really
distinguish the connections trying to engage with her (yay, pun), and a new
unique source ID is assigned to an engaged adapter, there is no limit on how
many adapters can be active with a single Priscilla server at any time. You can
effectively have both HipChat, Slack, IRC, as well as a local shell test console
all connect to the same Priscilla server the same time, or connect multiple
HipChat/Slack/IRC organizations through multiple copies of the same adapters
to the same Priscilla Server.

Currently only one adapter is functional, the HipChat adapter:
https://github.com/priscillachat/priscilla-hipchat

Before Priscilla server would recognize the adapter, adapter has to first
"engage" the Priscilla server with engagement command. This would cause
Priscilla server to check and either confirm or assign a newly generated unique
source identifier for the message source. No messages would be forwarded if
engagement does not succeed.

### Priscilla Responder

Another unique feature about Priscilla is that because it communicates with
other components via JSON over TCP, there isn't any requirement that
Priscilla's plugin has to be written in go. Any program/process that can speak
Priscilla's communication protocol can act as an adapter or responder. In fact,
there's an entire subclass of responders called passive responder, does not
even require the responder to speak Priscilla's protocol at all. They are
literally commands Priscilla server would execute on the fly upon deteting a
matching pattern from the incoming message from an adapter and return the
output of the command back to the origin adapter. For example, the ping command
is simply implemented as a passive responder using unix "echo" command. (See
conf-example.yml)

There are two types of Responders **Active Responder** and **Passive Responder**

#### Active Responder

Like adapter, active responder is a long running process that listens for
requests form Priscilla server. And like adapter, active responder has to first
engage the Priscilla server. Then active responder would perform regex pattern
registration, where it tells Priscilla server what message should be forwarded to
it. (This is not currently functional yet)

#### Passive Responder

This is a unique feature of Priscilla. A passive responder is essentially an
executable command that is specified in the Priscilla's config file what message
patter would trigger its execution and how it's executed. Once triggered, the
payload specified by config would be passed in as parameters, then the output of
the command would be captured and returned to the source adapter that triggered
the responder.

This is a powerful feature that makes both implementing and testing Priscilla
responders easy. You can literally write a single executable command that takes
input and produces output, and test it all without the need to have Priscilla
server running. When it's ready to be connected to Priscilla server, you
only have to specify in the config file how you want the command to be invoked.

For example, this is how you would implement a "ping" passive responder using
unix "echo" command, simply put in your Priscilla config file:

```yaml
responders:
  passive:
  - name: ping        # this is returned to adapter as source of the message
    match:
    - "^ping$"        # regular expression, you can specify multiple
    cmd: /bin/echo    # the command to be executed
    args: ["Pong"]    # the argument passed to the command
```

Passive responder can facilitate substring match substitution up to 10
submatches (0-9) enclosed in double-underscores. For example:

```yaml
responders:
  passive:
  - name: echo
    match:
    - "^echo (.+)$"
    cmd: /bin/echo
    args: ["__0__"]   # __0__ will be substituted by first submatch
```

Do be careful using the substitution, as it may have security concern. I would
recommend running Prescilla in a jailed environment (i.e. docker) to prevent
excape.

## Configuration

The configuration file is in YAML format, and you would specify the
configuration file with **-conf** argument when starting Priscilla server.

```yaml
port: 4517    # default port for Priscilla server
prefix: pris  # default prefix
prefix-alt: [priscilla $mention cilla] # alternate prefix, not yet implemented
adapters:     # adapter auto launch, not yet implemented
  hipchat:
    exec: /vagrant/dev/golang/bin/priscilla-hipchat
    params:
      user: "priscilla@priscilla.chat"
      pass: "abcdefg"
      nick: "Priscilla"
responders:
    passive:
    - name: echo
      match:
      - ^ping$
      noprefix: false # can be omitted, default is no activation without prefix
      cmd: /bin/echo
      args: ["pong"]
      fallthrough: false # can be omitted, if this is set to true, it will
                         # continue to match other patterns for activation,
                         # default behavior is to stop checking once it's
                         # activated once
    - name: cleverbot
      match:
      - ^(.*), pris$  # multiple activation patterns
      - ^(.*), pris\?$
      mentionmatch:   # match these patterns if it's being explictly mentioned
      - ^(.*)$
      noprefix: true  # will activate without prefix
      cmd: /usr/priscilla-scripts/cleverbot.sh # a script to curl cleverbot
      args: ["__0__"] # substitute with first submatch
    - name: sha256
      match:
      - ^sha256 (.+)$
      cmd: /usr/priscilla-scripts/sha256.sh
      args: ["__0__"]
        # content of sha256.sh:
        # #!/bin/bash
        #
        # echo $1 | shasum -a 256 | cut -f 1 -d ' '
```

## Some background

Priscilla is a chat bot written purely in go. It all started when I started
learning go and wanted to do something useful and practical (um...chat bot is
useful and practical...right?) so I can practice the newly acquired go
knowledge. I tried to be as idiomatic go as this is a learning experience for me
to use go. I tried really hard to make the code free of locks and use the *share
memory by communicating* methodology. So far it's successfuly. Though that's not
to say I won't eventually stumble onto a problem that I couldn't solve using
this methodology. But at the moment, the entire Priscilla code base is
mutex-free and lock-free. If you're to write a new Priscilla adapter or
responder, I would recommend you to do the same, since mixing locks and channels
could increase the chance of getting deadlocks (reference: ??? I know I've read
about it somewhere, I just have to track down the article...)

At where I work, we make good use of our chat bot which is Lita bot
(https://www.lita.io/) based. I've written quite a few Lita plugins, some
are custom internal plugins strictly used within the company, some got published
as open source plugin. Before the Lita bot, we had a hubot, which I too
contributed to some internal plugins, but everybody in the team was tired of
writting coffeescript so it didn't take long before it was abondanded after the
lita bot came online.

Having worked with both Hubot and Lita bot, I've come to know them quite a bit
about them, including some of the internals. One thing common about both Hubot
and Lita bot as well several other bots out there is that they all work in a
"include" model, meaning when you want to run a bot, you essentially take a
copy of the bot code and add more custom code to it so they are "included" in
the new instance of the bot code. This model work pretty well with interpreted
languages because you always get a copy of the source code when you run anything
in the interpreted language. You're really not losing anything, nor adding any
overhead with the model.

One shortfall I can think of, though, is the fact it forces a very particular
way you can implement your bot's plugin. It forces not only communication
protocol, but also the implementation details.

With a compiled language, though, this model does not work well at all. As I
mentioned earlier that if the a bot written in go is to follow the same
"include" model, then you'll have to re-compile your code every time you make
modifications to your bot to add functionality. I think for majority of the
administrator of chat bots, that's something they'd shy away from. I would. To
get around the problem, I developed a model where routing is detached from the
message generator, so that all components can be individually seleted and
assembled and change in one wouldn't need to affect the change in another (as
long as the communication protocol stays the same).

## Communication

* (A) Adapter
* (R) Responder
* (S) Priscilla Main Server

### Adapter engage (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "engage",
		"type": "adapter",
		"time": unixtimestamp,
		"data": "sha256-HMAC(unixtimestamp+source_identifier+secret)"
	}
}
```

### Active responder engage (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "engage",
		"type": "responder"
		"time": unixtimestamp,
		"data": "sha256-HMAC(unixtimestamp+source_identifier+secret)"
	}
}
```

**Note**

Other than the "type" field, adapter and responder engagement messages are
identical. However, Priscilla server treats adapter and responder differently.
Messages from Adapters, if a "to" field is ever set, it will be ignored as they
will always go through the pattern matching routine in the dispatcher. Messages
from responder on the other hand, would never go through a pattern matching
routine and would be sent directly to the "to" indicated source.

Timestamp must be within 10 seconds of server's current time. If the timestamp
is within range and HMAC matches the server's calculation, a "proceed" command
will be sent back as acknowledgement. If either time is out of range, or HMAC
calculation doesn't match, a "terminate" command will be sent back followed by
closing of the connection.

### Engagement success response (S->A, S->R)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "proceed",
		"data": "source_identifier",
	}
}
```

### Engagement failure response (S->A, S->R)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "terminate",
		"data": "error message",
	}
}
```

**Note**

When responder or adapter engage with Priscilla server, after the initial
engagement command is sent, client should wait for a command query. If it's an
command with action "proceed", then engagement is confirmed and normal
activities can begin. The source identifier that is assigned to the source will
be returned as the value in the "data" field. If the client's source id has
been accepted, it will be returned in this field. If it has been assigned a new
source identifier, it too will be returned in this field. Though this has very
little use for client as the server dispatcher would always ignore the source
field after initial engagement and insert the source id from the initial
engagement negation to every query came from that source.

If an error occures during the engagement process, then server would send back
a "terminate" command with an error message as the value in the "data" field,
then close the connection afterward.


### Disengage request (S->R/A, R/A->S)

```json
{
	"type": "command",
	"source": "identifier_or_server",
	"command": {
		"id": "identitfier",
		"action": "disengage",
	}
}
```

### Active responder command registration (R->S)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "register",
		"type": "pattern",
		"data": "regex_string"
		"options": ["fallthrough", "unhandled"]
	}
}
```

### Active responder command registration complete notification (R->S)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "noop",
		"type": "command",
	}
}
```

### Message from adapter (A->S)

**note:** "to" field can be left empty

**note:** "stripped" field is necessary because Adapters keeps it's own
reference of the mention name and it's not passed to the server. Server would
not know how to strip out the mention reference if it needs a stripped copy of
the message.

```json
{
	"type": "message",
	"source": "source_identifier",
  "to": "server",
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier",
    "mentioned": boolean,
    "stripped": "message stripped of mentions"
	}
}
```

### Message from active responder (R->S)

```json
{
	"type": "message",
	"source": "source_identifier",
  "to": "dest_identifier",
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier",
    "mentionnotify": ["user1", "user2", "user3"]
	}
}
```

### Request user information (S->A)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "request",
		"type": "user",
		"options": ["user1", "user2", "user3", "user4"]
	}
}
```

### User information response (A->S)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "info",
		"type": "user",
		"array": ["data1", "data2", "data3", "data4"]
	}
}
```

### Request room information (S->A)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "request",
		"type": "room",
		"options": ["room1", "room2", "room3", "room4"]
	}
}
```

### Room information response (A->S)

**Not Yet Implemented**

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "info",
		"type": "room",
		"array": ["data1", "data2", "data3", "data4"]
	}
}
```

