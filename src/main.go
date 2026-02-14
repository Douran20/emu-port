package main

import (
	_ "embed"
	lo "emu-port/src/logic"
	"image/color"
	"math" // i cannot be trusted
	"os"
	"github.com/gen2brain/raylib-go/raylib"
)

const DEBUG bool = false

//go:embed resource/HackNerdFontMono-Regular.ttf
var fontData []byte

func main() {
	config_buffer := &lo.Config{}
	lo.ReadJsonConfigs(config_buffer)

	var game_process *lo.GameProcess
	
	var buffer []lo.GameBuffer
	for i := 0; i < len(config_buffer.Game_pathG); i++ {
		var tempBuffer []string
		lo.ScanDir(&tempBuffer, "/"+config_buffer.Game_pathG[i], config_buffer.ExtensionG[i])
	
		for _, path := range tempBuffer {
			matches := lo.RegexName(path, config_buffer.RegexG[i])
			
			buffer = append(buffer, lo.GameBuffer{
				GamePlatform: config_buffer.PlatformG[i], 
				GamePath: path, 
				GameBin: config_buffer.BinG[i],
				GameArgs: config_buffer.ArgsG[i],
				GameName: matches+" - "+config_buffer.PlatformG[i],})
		}
	}

	if !DEBUG {rl.SetTraceLogLevel(rl.LogError)}
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagMsaa4xHint)
	rl.InitWindow(1920, 1080, "emu-port")
	
	font := rl.LoadFontFromMemory(".ttf", fontData, 64, nil)
	defer rl.UnloadFont(font)
	rl.HideCursor()
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)

	var flick_stick bool = true

	var inputtimer float32
	var selectIndex int = len(buffer)/2
	var virutalSelectindex float32 = 0

	rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {
		if rl.IsGamepadButtonPressed(0, rl.GamepadButtonMiddleRight) && game_process == nil {
			rl.CloseWindow()
			os.Exit(0)
		}

		var delta_time float32 = rl.GetFrameTime()
		if (inputtimer > 0) {inputtimer -= rl.GetFrameTime()}

		joystick_y := rl.GetGamepadAxisMovement(0, rl.GamepadAxisLeftY)
		virutalSelectindex = rl.Lerp(virutalSelectindex, float32(selectIndex), 7.0 * delta_time)

		if game_process != nil {
			print(game_process.Buffer.String())
			if lo.ShouldGameDie(game_process) {
				game_process = nil
			}
		}

		rl.BeginDrawing()
		rl.DrawRectangleGradientV(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), color.RGBA{12, 9, 51, 255}, color.RGBA{53, 30, 99, 255})
		if game_process != nil {
			rl.ClearBackground(rl.Black)
		} else if game_process == nil {
			if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceDown) || rl.IsKeyPressed(rl.KeyDown) {selectIndex++}
			if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceUp) || rl.IsKeyPressed(rl.KeyUp)  {selectIndex--} 

			if inputtimer <= 0 && math.Abs(float64(joystick_y)) > 0.5 {
				if (joystick_y < -0.5) {selectIndex--} else {selectIndex++}
				if flick_stick {
					inputtimer = 0.6
					flick_stick = false
				} else {
					inputtimer = 0.05
				}
			}	

			if math.Abs(float64(joystick_y)) < 0.2 {
				inputtimer = 0
				flick_stick = true
			}

			if selectIndex < 0 {selectIndex = len(buffer) -1}
			if selectIndex >= len(buffer) {selectIndex = 0}

			for i, file := range buffer {		
				var item_offset_y float32 = (float32(i)*150)+(float32(-virutalSelectindex)*150)+float32(rl.GetScreenHeight()/2-80)			
				var item_offset_x float32 = float32(rl.GetScreenWidth()/8) 
				var rec_width float32 = float32(len(file.GameName)+2)*(float32(rl.GetScreenWidth())*0.0145)
				item_vector := rl.Vector2{X: float32(item_offset_x+12), Y:float32(item_offset_y)} 		
				border_rec := rl.Rectangle{X: float32((item_offset_x-20)), Y: float32(item_offset_y-20), Width: rec_width, Height: 100.0}
				scrollbar_rec := rl.Rectangle{X: float32(rl.GetScreenWidth()-20), Y: ((virutalSelectindex)*float32(rl.GetScreenHeight()/len(buffer))), Width: 10, Height: (float32(rl.GetScreenHeight()/len(buffer)))}
				
				if selectIndex == i {
					border_rec.X += 40
					item_vector.X += 40
					rl.DrawRectangleRounded(border_rec, 0.5, 8, color.RGBA{35, 45, 69, 255})
					rl.DrawRectangleRoundedLinesEx(border_rec, 0.5, 8, 2.0, color.RGBA{159, 144, 192, 255})
					rl.DrawTextEx(font, file.GameName, item_vector, 50, 1, color.RGBA{188, 162, 242, 255})
				} else {
					rl.DrawRectangleRounded(border_rec, 0.2, 8, color.RGBA{19, 24, 38, 255})
					rl.DrawTextEx(font, file.GameName, item_vector, 50, 1, color.RGBA{74, 57, 110, 255})
					rl.DrawRectangleRounded(scrollbar_rec, 1, 8, color.RGBA{255, 255, 255, 255})
				}
			}
			if rl.IsGamepadButtonPressed(0, rl.GamepadButtonRightFaceDown) {
				println("X button Pressed")
				selected := buffer[selectIndex]

				var args []string

				for _, a := range selected.GameArgs {
					if a == "$" {
						println(selected.GamePath)
						args = append(args, selected.GamePath)
					} else {
						println(a)
						args = append(args, a)
					}
				}
				
				game, err := lo.RunGame(buffer[selectIndex].GameBin, args)
				println("launched process ", game)
				//println(args)
				if err == nil {
					game_process = game
				}
			}
			rl.ClearBackground(rl.Black)
		}
		rl.EndDrawing()
	}
}
