# Where should we write the GIF?
Output demo.gif

# Set up a 1200x600 terminal with 46px font.
Set FontSize 32
Set Width 1200
Set Height 600

Set Margin 20
Set MarginFill "#674EFF"
Set BorderRadius 10

#setup# build tuido executable
Hide
Type "go build -o tuido ."
Enter
Show
#/setup#

Type "./tuido"
Sleep 500ms
Enter

Sleep 2

Down 3
Up 1
Down 5
Tab@2s 2 # show done, back to todo

# filtering
Type "/"
Type "#feat"
Down
Up

# peek the feat todo
Enter@2s 2
type "/"

# clear the filter
Backspace 5

# show a pomodoro (currently broken! launching pomo also launches peek - does not capture 'enter')

Up 3
Type "p"
Sleep 350ms
type "0.1"
Enter

# clear todos via done / sleep / strike

Type "x"
Down
Type "s"
Down
Type "a"
Down
Type "z"
Down
