package main

type calculate struct {
	name      string           // Name of the file
	total     int              // Total line of the file
	inlineSet map[int]struct{} // Set of integers for inline calculation
	blockSet  map[int]struct{} // Set of integers for block calculation
}

type calState struct {
	isInblock       bool // Indicates if the state is inside a block
	isInquotes      bool // Indicates if the state is inside double quotes
	isInSingleQuote bool // Indicates if the state is inside single quotes
	isGotoEnd       bool // Indicates if the state should pass the line
	inRMode         int  // Indicates the mode of the state for R string
}
