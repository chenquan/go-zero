package version

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// GoZeroVersion go-zero version
	GoZeroVersion = "1.3.1"
	banner        = `
________  ________               ________  _______   ________  ________     
|\   ____\|\   __  \             |\_____  \|\  ___ \ |\   __  \|\   __  \    
\ \  \___|\ \  \|\  \  ___________\|___/  /\ \   __/|\ \  \|\  \ \  \|\  \   
 \ \  \  __\ \  \\\  \|\____________\ /  / /\ \  \_|/_\ \   _  _\ \  \\\  \  
  \ \  \|\  \ \  \\\  \|____________|/  /_/__\ \  \_|\ \ \  \\  \\ \  \\\  \ 
   \ \_______\ \_______\            |\________\ \_______\ \__\\ _\\ \_______\
    \|_______|\|_______|             \|_______|\|_______|\|__|\|__|\|_______|
`
)

func PrintGoZeroVersion() {
	fmt.Print(banner)
	logx.Infof("go-zero version:%s", GoZeroVersion)
}
