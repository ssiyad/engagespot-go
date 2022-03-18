![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
# engagespot-go
Go SDK for [Engagespot](https://engagespot.co/)

### Installation
```shell
go get github.com/ssiyad/engagespot-go
```

### Example
```go
package main

import "github.com/ssiyad/engagespot-go"

func main() {
    c := engagespot.NewEngagespotClient("API_KEY", "API_SECRET")

    n, _ := c.NewNotification("Hello world!")
    n.SetMessage("Let's go!")
    n.SetIcon("https://example.com/icon.svg")
    n.SetUrl("https://example.com")
    n.SetCategory("greet")
    n.AddRecipient("hello@example.com")

    n.Send()
}
```
