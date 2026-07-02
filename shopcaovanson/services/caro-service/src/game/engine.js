const DIRECTIONS = [
  [0, 1],
  [1, 0],
  [1, 1],
  [1, -1],
]

export function createBoard(size) {
  return Array.from({ length: size }, () => Array(size).fill(null))
}

export function cloneBoard(board) {
  return board.map((row) => [...row])
}

export function inBounds(size, row, col) {
  return row >= 0 && row < size && col >= 0 && col < size
}

export function countLine(board, size, row, col, dr, dc, symbol) {
  let count = 0
  let r = row
  let c = col
  while (inBounds(size, r, c) && board[r][c] === symbol) {
    count += 1
    r += dr
    c += dc
  }
  return count
}

/** Exactly 5 in a row wins; 6+ (overline) does not count. */
export function checkWin(board, size, row, col, symbol) {
  for (const [dr, dc] of DIRECTIONS) {
    const total =
      countLine(board, size, row, col, dr, dc, symbol) +
      countLine(board, size, row - dr, col - dc, -dr, -dc, symbol) -
      1
    if (total === 5) return true
  }
  return false
}

export function isBoardFull(board) {
  return board.every((row) => row.every((cell) => cell !== null))
}

function countOpenThrees(board, size, symbol) {
  let openThrees = 0
  for (let r = 0; r < size; r++) {
    for (let c = 0; c < size; c++) {
      if (board[r][c] !== symbol) continue
      for (const [dr, dc] of DIRECTIONS) {
        if (lineLength(board, size, r, c, dr, dc, symbol) === 3 && isOpenThree(board, size, r, c, dr, dc, symbol)) {
          openThrees += 1
        }
      }
    }
  }
  return openThrees
}

function lineLength(board, size, row, col, dr, dc, symbol) {
  let len = 0
  let r = row
  let c = col
  while (inBounds(size, r, c) && board[r][c] === symbol) {
    len += 1
    r += dr
    c += dc
  }
  return len
}

function isOpenThree(board, size, row, col, dr, dc, symbol) {
  const len = lineLength(board, size, row, col, dr, dc, symbol)
  if (len !== 3) return false
  const endR = row + dr * len
  const endC = col + dc * len
  const startR = row - dr
  const startC = col - dc
  const startOpen = inBounds(size, startR, startC) && board[startR][startC] === null
  const endOpen = inBounds(size, endR, endC) && board[endR][endC] === null
  return startOpen && endOpen
}

/** Simplified double-three block: reject if move creates 2+ open threes for mover. */
export function isForbiddenMove(board, size, row, col, symbol) {
  const next = cloneBoard(board)
  next[row][col] = symbol
  return countOpenThrees(next, size, symbol) >= 2
}

export function validateMove(board, size, row, col, symbol, enforceDoubleThree = true) {
  if (!inBounds(size, row, col)) return 'out of bounds'
  if (board[row][col] !== null) return 'cell occupied'
  if (enforceDoubleThree && isForbiddenMove(board, size, row, col, symbol)) {
    return 'forbidden: double-three'
  }
  return null
}

export function resolveRps(a, b) {
  if (a === b) return 'tie'
  if (
    (a === 'rock' && b === 'scissors') ||
    (a === 'paper' && b === 'rock') ||
    (a === 'scissors' && b === 'paper')
  ) {
    return 'a'
  }
  return 'b'
}
