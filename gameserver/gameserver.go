package gameserver

import(
    "hema/hangman/rpc/hangmantypes"
    "errors"
    "net"
    "net/rpc"
    "fmt"
    "strconv"
    "hema/hangman/rpc/asyncio"
)


type PlayerId uint16

type PlayerAddress struct{
    Ip string
    Port string
}


type Player struct{
    Pid PlayerId
    IsHost bool
    Addr PlayerAddress
    Client *rpc.Client
    //Client PlayerClient
}

func InitPlayer(id PlayerId, addr PlayerAddress)( Player, error){
    
    var player Player
    
    client, err := rpc.Dial("tcp", addr.Ip + ":" + addr.Port)

    if err != nil{
        fmt.Println(err)
        return player, errors.New("Could not connect to Client")
    }
    player.Pid = id
    player.IsHost = (id == 0)
    player.Addr = addr
    player.Client = client

    return player, nil
}

type PlayerMetadata struct{
    PlayerCount int
    Players [] PlayerId
}

type HangmanGameServer struct {
    H_game hangmantypes.HangmanGame
    PlayerCount int
    Players [] Player
    LastTurn PlayerId
}

func (h_gameserver *HangmanGameServer) GetNewPid() (PlayerId, error){

   for i := 0; i < h_gameserver.PlayerCount; i++ {
        _, err := h_gameserver.GetPlayerByPid(PlayerId(i))
        
        if err != nil{
            return PlayerId(i), nil
        }

   }

    return PlayerId(h_gameserver.PlayerCount), nil

}

func (h_gameserver *HangmanGameServer) GetPlayerByPid(pid PlayerId) (*Player, error){

    for _, p := range h_gameserver.Players{
        if p.Pid == pid{
            return &p, nil
        }
    }

    return nil, errors.New("GetPlayerByBid: No player with pid" + strconv.Itoa(int(pid)))
}

func (h_gameserver *HangmanGameServer)Join(addr *PlayerAddress, pid *PlayerId) error {
   
    var err error

    *pid, err = h_gameserver.GetNewPid()  
    //create new player Client
    player, err := InitPlayer(*pid, *addr)

    if(err != nil){
        fmt.Println(err)
        return err
    }
    
    //add the player to the list on success
    h_gameserver.Players = append(h_gameserver.Players, player)

    //register player id as return value
    h_gameserver.PlayerCount++

    var tmp uint8
    var player_info PlayerMetadata
    player_info.PlayerCount = h_gameserver.PlayerCount
    player_info.Players = make([]PlayerId, player_info.PlayerCount)
    for i, p := range h_gameserver.Players{
        player_info.Players[i] = p.Pid
    }

    for _, p := range h_gameserver.Players{
        fmt.Println("Hi?")
        err = p.Client.Call("HangmanPlayerServer.UpdateLobby", &player_info, &tmp)
    
        if err != nil{
            fmt.Println(err)
        }
    
    }

    fmt.Println("A new player joined! (Player", *pid, ")")

    return nil

}

func (h_gameserver *HangmanGameServer)Quit(pid *PlayerId, result *bool) error{

    fmt.Println("Player", *pid, "quit the lobby")

    var p *Player = nil

    p, err := h_gameserver.GetPlayerByPid(*pid)

    if(err != nil){
        return err
    }

    err = p.Client.Close()
    if err != nil {
        return err
    }
    
    for i, player := range h_gameserver.Players{
        if(i != h_gameserver.PlayerCount-1 && player.Pid == *pid){
            h_gameserver.Players[i] = h_gameserver.Players[h_gameserver.PlayerCount-1]
            break
        }
    } 
    h_gameserver.PlayerCount--

    h_gameserver.Players = h_gameserver.Players[:h_gameserver.PlayerCount]

    var tmp uint8


    var player_info PlayerMetadata

    player_info.PlayerCount = h_gameserver.PlayerCount
    player_info.Players = make([]PlayerId, player_info.PlayerCount)

    for i, p := range h_gameserver.Players{
        player_info.Players[i] = p.Pid
    }
    for _, p := range h_gameserver.Players{
        fmt.Println("Updating!")
        err = p.Client.Call("HangmanPlayerServer.UpdateLobby", &player_info, &tmp)

        if err != nil{
            fmt.Println(err)
        }
    }

    return nil
}

func (h_gameserver *HangmanGameServer)StartGame(pid *PlayerId, result *bool) error {
    

    var tmp1, tmp2 uint8 

    for _, p := range h_gameserver.Players{
        p.Client.Call("HangmanPlayerServer.StartGame", &tmp1, &tmp2)
    }

    h_gameserver.H_game.Init()

    h_gameserver.Run()

    return nil
}

func GetPlayerString(p Player, pid PlayerId, done chan PlayerId, error_chan chan error, res *asyncio.IoResponse){
    err :=p.Client.Call("HangmanPlayerServer.GetString", &pid, res)
    if err != nil{
        fmt.Println("Error in GetPlayerString:", err)
    }

    done <- p.Pid
    error_chan <- err

    return 

}

func (h_gameserver *HangmanGameServer) ChooseString() (asyncio.IoResponse, error){

    pid := h_gameserver.LastTurn

    fake_res := make([]asyncio.IoResponse, h_gameserver.PlayerCount)
    var res asyncio.IoResponse


    done := make(chan PlayerId)
    error_chan := make(chan error)

    var err error

    for i, p := range h_gameserver.Players{
        if(p.Pid == pid){
            go GetPlayerString(p, pid, done, error_chan, &res) 
        }else{
            go GetPlayerString(p, pid, done, error_chan, &fake_res[i]) 
        }
    }

    for i := 0; i < h_gameserver.PlayerCount; i++{
        if <-done == pid {
            err = <- error_chan
            if(err != nil){
                return res, err 
            } 
            return res, nil
        }
    }

    return res, errors.New("ChooseString(): No Player with this pid")
    

}

func (h_gameserver *HangmanGameServer)Run() error {

    var response asyncio.IoResponse
    var err error

    h_gameserver.LastTurn = PlayerId(0);

    for{

        response, err = h_gameserver.ChooseString()

        if response.Canceled == false {
            h_gameserver.H_game.H_str.Init(response.Result)
            break
        }


        _, _ = h_gameserver.ChooseNextPlayer()

    }

    for !h_gameserver.H_game.Over() {

        h_gameserver.ShareState()

        err = h_gameserver.NextTurn()

        if err != nil{
            fmt.Println("Error during Run:", err)
            return nil
        }
    
    }
    

    h_gameserver.ShareState()

    return nil
}


func (h_gameserver *HangmanGameServer)ChooseNextPlayer() (PlayerId, error){

    
    for i, p := range h_gameserver.Players{
        if p.Pid == h_gameserver.LastTurn{
            if i == h_gameserver.PlayerCount - 1{
                h_gameserver.LastTurn = h_gameserver.Players[0].Pid
                return h_gameserver.LastTurn, nil
            }else{
                h_gameserver.LastTurn = h_gameserver.Players[i+1].Pid
               return h_gameserver.LastTurn, nil
            }
        }
    }


    return PlayerId(9999), errors.New("ChooseNextPlayer: Weird Error Occurred")


}
    
func GetPlayerTurn(p Player, pid PlayerId, done chan PlayerId, error_chan chan error, res *asyncio.IoResponse){
    err :=p.Client.Call("HangmanPlayerServer.DoTurn", &pid, res)
    if err != nil{
        fmt.Println("Error in GetPlayerTurn:", err)
    }

    done <- p.Pid
    error_chan <- err

    return 

}

func (h_gameserver *HangmanGameServer)NextTurn() error{

    for{
        pid, err := h_gameserver.ChooseNextPlayer()

        if err!= nil{
            fmt.Println(err)
            return err
        }

        fake_res := make([]asyncio.IoResponse, h_gameserver.PlayerCount)
        var res asyncio.IoResponse

        done := make(chan PlayerId)
        error_chan := make(chan error)

        for i, p := range h_gameserver.Players{
            if p.Pid == pid {
                go GetPlayerTurn(p, pid, done, error_chan, &res) 
            }else{
                go GetPlayerTurn(p, pid, done, error_chan, &fake_res[i]) 
            }
        }
        for i := 0; i < h_gameserver.PlayerCount; i++{
            if <-done == pid {
                err = <- error_chan
                if(err != nil){
                    return err 
                } 

                if res.Canceled == true {
                    continue
                }

                h_gameserver.H_game.DoTurn(res.Result)
                return nil
            }
        }
    }
return nil



}

func (h_gameserver *HangmanGameServer)ShareState() error {

    res := make([]bool, h_gameserver.PlayerCount)
   

    for i, p := range h_gameserver.Players{
        err := p.Client.Call("HangmanPlayerServer.UpdateDisplay", &h_gameserver.H_game, &res[i])
            if err != nil{
                fmt.Println(err)
                return err
            }
    }

    return nil

}
func Init() (*HangmanGameServer, error){

    h_gameserver := new(HangmanGameServer)
    
    h_gameserver.PlayerCount = 0

    handler:= rpc.NewServer()

    err := handler.Register(h_gameserver)

    if err != nil{
        fmt.Println(err)
        return h_gameserver, err
    }

    handler.HandleHTTP("/HangmanGameServer", "/debug/HangmanGameServer")

    listener, err := net.Listen("tcp", "0.0.0.0:1234")

    if err != nil{
        
        fmt.Println(err)

        return h_gameserver, errors.New("Failed to Listen")
    }

    go func() {
        for{
            cxn, _ := listener.Accept()
            go handler.ServeConn(cxn)
        }
    }()

    return h_gameserver, nil

}
