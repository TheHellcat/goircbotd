ircbotd - supplemental IRC bot daemon
==========

**This project is an IRC bot daemon written in Go (Golang) that works as interface between the actual IRC bot project (written in PHP/Symfony) and the IRC server(s)**

It implements the persistent connection to the IRC server, handling of the IRC protocol and all communication with the server.  
The bot daemon can not run on it's own, it's only meant as an interface between the PHP project and the IRC server as PHP is - by design - not meant for certain tasks, like keeping up persistent connections and running as background process over a (very) long time.

The bot daemon fetches it's config via HTTP request from the parent project, handles all communication with the IRC server and passes all recognized/registered chat-commands (incl. timed ones set via config) to the PHP project to then pass PHPs response back to the IRC server.

All actual logic for chat commands and similar things is implemented in the PHP side project.
