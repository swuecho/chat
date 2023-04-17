/**
 * Generates a random string of length n
 * @param n Length of the random string
 */
export function generateRandomString(n: number): string {
  // Array of possible characters
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  // Random string to be returned
  let randomString = ''

  // Loop n times to generate a string of length n
  for (let i = 0; i < n; i++) {
    // Get a random character from the array of characters
    const randomCharacter = characters.charAt(Math.floor(Math.random() * characters.length))
    // Append the random character to the random string
    randomString += randomCharacter
  }
  return randomString
}
