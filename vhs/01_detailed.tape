Output demos/01_detailed.gif

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 600
Set Padding 16
Set Framerate 15
Set TypingSpeed 30ms

Type "kdn info"
Enter
Wait+Screen@5000ms /Runtimes:/
Sleep 5000ms

Type "kdn init --runtime podman --agent claude --name demo"
Enter
Wait+Screen@5000ms /Container created/
Sleep 5000ms

Type "kdn list"
Enter
Wait+Screen@5000ms /demo/
Sleep 5000ms

Type "kdn start demo"
Enter
Wait+Screen@5000ms /Container started/
Sleep 5000ms

Type "clear"
Enter
Sleep 1500ms

Type "kdn terminal demo"
Enter
Sleep 5000ms

Type "/exit"
Enter
Sleep 5000ms

Type "clear"
Enter
Sleep 1500ms

Type "kdn stop demo"
Enter
Wait+Screen@20000ms /Container stopped/
Sleep 5000ms

Type "kdn remove demo"
Enter
Wait+Screen@5000ms /Container removed/
Sleep 5000ms

Type "kdn list"
Enter
Wait+Screen@5000ms /No workspaces registered/
Sleep 5000ms
