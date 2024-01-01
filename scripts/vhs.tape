# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type@1ms:
#   Type@1ms[@<time>] "<characters>"    Type@1ms <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output "ignore.gif"
Require wander

Set Padding 10
Set FontSize 13
Set Width 1100
Set Height 450
Set Framerate 15

Type@1ms "wander"
Enter

# All Jobs
Type@1ms "j"
Sleep 200ms
Screenshot ../img/home.png
Sleep 200ms

# All Events
Type "V"
Sleep 200ms
Screenshot ../img/screenshots/Global_Events.png
Sleep 200ms

# Job Events
Escape 2
Type@1ms "v"
Sleep 200ms
Screenshot ../img/screenshots/Job_Events.png
Sleep 200ms

# All Tasks
Type@1ms "A"
Type@1ms "jj"
Type@1ms "k"

# Exec for Task
Type@1ms "e"
Backspace 10
Type@1ms "bash"
Enter
Type@1ms "ls -la"
Enter
Ctrl+d
Escape 2

# Back to All Jobs
Type@1ms "J"

# Tasks for Job
Enter

# Logs for Task, Filtering
Enter
Type@1ms "/"
Type@1ms "INFO"
Enter
Type@1ms "nnnN"
Type@1ms "jjkk"
Enter
Escape 2

# Save Logs
Ctrl+S
Type@1ms "my_logs.txt"
Enter

# Back to Tasks for Job
Escape

# Stats for Allocation
Type@1ms "s"

# Back to Tasks for Job
Escape
