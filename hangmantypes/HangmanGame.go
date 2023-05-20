package hangmantypes

import(
    "fmt"
    "strconv"
    "unicode"
    "hema/hangman/rpc/hangmandisplay"
)



type HangmanGame struct {
    Stage uint8
    Symbol [10] string
    Log [3] string
    H_str HangmanString
}

func (h_game *HangmanGame)Init(){
    h_game.Stage = 0

    h_game.Symbol[0] = "       \n        \n        \n        \n        \n" 
    h_game.Symbol[1] = "       \n  |     \n  |     \n  |     \n  |     \n" 
    h_game.Symbol[2] = "   __  \n  |     \n  |     \n  |     \n  |     \n" 
    h_game.Symbol[3] = "   __  \n  |  |  \n  |     \n  |     \n  |     \n"
    h_game.Symbol[4] = "   __  \n  |  |  \n  |  O  \n  |     \n  |     \n"
    h_game.Symbol[5] = "   __  \n  |  |  \n  |  O  \n  |  |  \n  |     \n"
    h_game.Symbol[6] = "   __  \n  |  |  \n  | \\O  \n  |  |  \n  |     \n"
    h_game.Symbol[7] = "   __  \n  |  |  \n  | \\O/ \n  |  |  \n  |     \n"
    h_game.Symbol[8] = "   __  \n  |  |  \n  | \\O/ \n  |  |  \n  | /   \n"
    h_game.Symbol[9] = "   __  \n  |  |  \n  | \\O/ \n  |  |  \n  | / \\ \n"

    h_game.Log[0] = ""
    h_game.Log[1] = ""
    h_game.Log[2] = ""


    //fmt.Print("Enter your string: ")

    //scanner.Scan()

    //h_game.H_str.Init(scanner.Text())

}


/*func (h_game *HangmanGame)Run(){
    
    hangmandisplay.PrintStart()

    for !h_game.Over()  {

        h_game.PrintState()


        h_game.DoTurn()

    }

    hangmandisplay.Clear()

    h_game.PrintState()

}*/



func (h_game *HangmanGame) InsertLog(str string){

    

    i := 0

    const max int = 3

    for i < max{
        if i == max-1 { h_game.Log[max-1] = str 
        }else          { h_game.Log[i] = h_game.Log[i+1] }
        i++
    }

}

func (h_game HangmanGame) PrintState(){
    
    hangmandisplay.Clear()

    fmt.Println("----HANGMAN----")

    fmt.Println(h_game.Symbol[h_game.Stage])

    fmt.Println("")

    h_game.H_str.PrintAll()

    fmt.Println("")

    fmt.Println("=============================")
    i := 0
    for i < 3{
        fmt.Println(h_game.Log[i])
        i++
    }
    fmt.Println("=============================")

    fmt.Println("")

}

func (h_game *HangmanGame) DoTurn(str string) {
    
        //fmt.Println("Your guess (one letter, or whole Word): ")

        //scanner := bufio.NewScanner(os.Stdin)

        //scanner.Scan()
        
        guess := str

        if len(guess) != 1 && len(guess) != len(h_game.H_str.Word) {
            h_game.InsertLog( "Invalid guess! Choose either one letter, or the entire Word.")
            return
        }

        if(len(guess) == 1 && !unicode.IsLetter(rune(guess[0]))){
            h_game.InsertLog( "Invalid guess! Only alphabetic characters as single guess.")
            return
        }

        count := h_game.H_str.ValidateGuess(guess)

        if(count == 0){
            h_game.Stage++
            h_game.InsertLog( "'" + guess + "' is not included...")
            return
        }

        h_game.InsertLog("Your guess '" + guess + "' revealed " + strconv.Itoa(int(count)) + " new letter(s)!")     

        return
}

func (h_game *HangmanGame) Over() bool{
    if h_game.Stage == 9{
        h_game.InsertLog("Game Over...")
        return true
    }

    if h_game.H_str.IsRevealed(){

        h_game.InsertLog("Congrats! You guessed the right Word!")
        return true
    }

    return false
}

