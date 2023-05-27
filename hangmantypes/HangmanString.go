package hangmantypes

import(
    "fmt"
    "strings"
)



type HangmanString struct{
    Word string
    Revealed [] bool
    L_arr LetterArray
}

func (h_str *HangmanString) Init(str string){
    
    h_str.Word = str

    h_str.Revealed = make([]bool,0)

    for _, char := range str {
        if ((char <= 'z' && char >= 'a') || (char <= 'Z' && char >= 'A')){
            h_str.Revealed = append(h_str.Revealed, false)
        } else{
            h_str.Revealed = append(h_str.Revealed, true)
        }
    }

    h_str.L_arr.Init()
}

func (h_str *HangmanString) ValidateChar(c uint8) uint{
 
    var count uint = 0
    for pos, char := range h_str.Word{
        if strings.ToLower(string(c)) == strings.ToLower(string(char)) && h_str.Revealed[pos] == false {
            h_str.Revealed[pos] = true;
            count++
        }
    }

    if count > 0 {
        h_str.L_arr.Update(c, true)
    }else{
        h_str.L_arr.Update(c, false)
    }

    return count

}


func (h_str *HangmanString) ValidateGuess(str string) uint{

    if len(str) != 1 && len(str) != len(h_str.Word){
        return 0
    }

    var count uint = 0

    for _, char := range str{
        count += h_str.ValidateChar(uint8(char))
    }

    return count

}

func (h_str HangmanString)IsRevealed() bool{

    for _, val := range h_str.Revealed{
        if val == false{
            return false
        }
    }

    return true
}

func (h_str HangmanString)PrintStr(){
    for pos, char := range h_str.Word{
        if h_str.Revealed[pos] == true{
            fmt.Print(string(char))
        }else{
            fmt.Print("_")
        }
    }
    fmt.Println("")
}

func (h_str HangmanString)PrintAll(){

    h_str.PrintStr()

    h_str.L_arr.PrintArr()

}
