package asyncio

import(
    "bufio"
    "os"
    "time"
)

type IoResponse struct{
    Canceled bool
    Result string
}

type MutexLock struct{
    lock chan int
}

func (mut *MutexLock)Lock(){
    <- mut.lock
}
func (mut *MutexLock)Unlock(){
    mut.lock <- 1
}
func (mut *MutexLock)Init(){
    mut.lock = make(chan int, 1)
    mut.lock <- 1
}


type IoRequest struct{
    Type int
    Canceled bool
    Completed bool
    Signal chan int
    Timeout int
    Response chan IoResponse
    mut MutexLock
}

type RunningIoRequest struct{
    Request *IoRequest
    Signal chan bool
}

type IoInstance struct{
    
    Input chan string
    Requests chan *IoRequest
    Queue chan *RunningIoRequest
}




func (io *IoInstance) Reader() error{
    reader := bufio.NewReader(os.Stdin)

    for{
        str, err := reader.ReadString('\n')
        if err != nil{
            return err
        }
        io.Input <- str
    }
}

func (io *IoInstance) RequestHandler() error{
    for{
        select{
        case req := <-io.Queue:
            select{
            case <-req.Signal:
                //IoRequest got canceled
                continue
            case str := <-io.Input:
                //IoRequest got completed
                req.Request.mut.Lock()
                if(req.Request.Canceled == true){
                    //this should be the only harmful scenario
                    //we consumed an input that was not able to be sent as a response
                    req.Request.mut.Unlock()
                    continue
                }
                req.Request.Completed = true
                req.Request.Canceled = false
                if(req.Request.Type == 2){
                    req.Request.Signal <- 1
                }
                var resp IoResponse
                resp.Canceled = false
                resp.Result = str[:len(str)-1]
                req.Request.Response <- resp
                req.Request.mut.Unlock()
            }
        case <-io.Input:
        }
    }
}

func RequestTimeoutHandler(running RunningIoRequest){

    

    time.Sleep(time.Duration(running.Request.Timeout) * time.Second)
    running.Request.mut.Lock()
    defer running.Request.mut.Unlock()
    if running.Request.Completed == true {
        return
    } 
    running.Signal <- true
    running.Request.Canceled = true
    var resp IoResponse
    resp.Canceled = true
    running.Request.Response <- resp

}

func RequestSignalHandler(running RunningIoRequest){
    
    <-running.Request.Signal 

    running.Request.mut.Lock()
    defer running.Request.mut.Unlock()
    if running.Request.Completed == true {
        return
    } 


    running.Signal <- true
    running.Request.Canceled = true
    var resp IoResponse
    resp.Canceled = true
    running.Request.Response <- resp

}

func (io *IoInstance) RequestScheduler() error{
    for{
        newReq := <-io.Requests 
    

        var running RunningIoRequest

        running.Request = newReq
        running.Signal = make(chan bool)

        if newReq.Type == 1 { //Timeout-based
            go RequestTimeoutHandler(running)
        }else if newReq.Type == 2 { //Interrupt-based
            go RequestSignalHandler(running)
        }

        io.Queue <- &running
    }
}

func NewIoInstance() (IoInstance, error){
    var io IoInstance

    io.Input = make(chan string)
    io.Requests = make(chan *IoRequest)
    io.Queue = make(chan *RunningIoRequest)

    return io, nil
}


func (io *IoInstance)Launch(){
    go io.RequestHandler()
    go io.Reader()
    go io.RequestScheduler()
}


func (io *IoInstance)RequestLine() IoResponse{

    var req IoRequest

    req.Type = 0 //blocking
    req.Canceled = false
    req.Completed = false
    req.Response = make(chan IoResponse)
    req.mut.Init() 


    io.Requests <- &req

    return <-req.Response

}

func (io *IoInstance)RequestLineTimeout(time int) IoResponse{

    var req IoRequest

    req.Type = 1
    req.Canceled = false
    req.Completed = false
    req.Timeout = time
    req.Response = make(chan IoResponse)
    req.mut.Init() 


    io.Requests <- &req

    return <-req.Response

}

func (io *IoInstance)RequestLineSignal(signal chan int) IoResponse{

    var req IoRequest

    req.Type = 2
    req.Canceled = false
    req.Completed = false
    req.Signal = signal
    req.Response = make(chan IoResponse)
    req.mut.Init() 
    io.Requests <- &req

    return <-req.Response
}


