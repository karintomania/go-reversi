# Go Reversi
Go revewsi is a CLI game to play reversi in terminal.  

Single Play (Local)
![screen recording](https://github.com/user-attachments/assets/ec7d106e-3daf-41a5-995a-c1cdc0e7bf05)

Online Play
![online](https://github.com/user-attachments/assets/de284b2a-8fc6-446e-bfd1-9b85d7233486)


# How to Install

## Docker (Recommened)
This command starts a game in single player mode with 8x8 board.  
```
docker run --rm -it ghcr.io/karintomania/go-reversi:latest 
```

See Examples section for more options.  

## Download Binary
Install the binary and run the file from [Releases Page](https://github.com/karintomania/go-reversi/releases).  

# Options
These are the options you can specify:  
```
  -h show help

  -n int
        Dimension of the board. (Default: 8) (default 8)

# For local play
  -p int
        1 for Single Play, 2 for 2 Players. (Default: 1)

# For online play
  -port int
        Specify game server's port (default 4696)

  -s    Start game as host. This option runs a game server

  -url string
        Start game as a client. This specifies the game server url to connect.

# For devlopment
  -d    Debug info
```

# Examples
## Local Play
Single Player mode with 8x8 board. (Change the command according to your platform and version)  
```
// with Binary
./go-reversi-0.1-linux-x86

// with Docker
docker run --rm -it ghcr.io/karintomania/go-reversi:latest
```

Local 2 Players mode with 6x6 board.  
```
// with Binary
./go-reversi-0.1-linux-x86 -n 6 -p 2

// with Docker
docker run --rm -it ghcr.io/karintomania/go-reversi:latest -n 6 -p 2
```

## Online Play
To play online, one player needs to run a game server, and another player connects to the server.  

Start a game server on port 4696:  
```
// with Binary
./go-reversi-0.1-linux-x86 -s

// with Docker
docker run --rm -it -p 4696 ghcr.io/karintomania/go-reversi:latest -s
```

Connect to a game server running on http://example.com:4696:  
```
// with Binary
./go-reversi-0.1-linux-x86 -url http://example.com

// with Docker
docker run --rm -it ghcr.io/karintomania/go-reversi:latest -url http://example.com
```

# Using ngrok
You can use a service like ngrok to temporalily publish your server.  

Currently, if you use `https` with ngrok, the docker version of this app shows error saying `x509: certificate signed by unknown authority`.  
You can avoid this by using `http` instead:
```
// On server side
docker run --rm -p 4696 -it ghcr.io/karintomania/go-reversi:latest -s
ngrok http 4696 --scheme http
// The output will look like this:
// Forwarding http://aaaa-00-00-000-000.ngrok-free.app -> http://localhost:4696

// On client side, you run this
docker run --rm -it ghcr.io/karintomania/go-reversi:latest -url http://aaaa-00-00-000-000.ngrok-free.app:80
```

# TODO
- [ ] Nice pop-up message
- [ ] Stronger AI
- [ ] Tidy code

