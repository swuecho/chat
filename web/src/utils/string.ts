export function getDataFromResponseText(responseText: string): string {
        // first data segment
        if (responseText.lastIndexOf('data:') === 0)
                return responseText.slice(5)
        // Find the last occurrence of the data segment
        const lastIndex = responseText.lastIndexOf('\n\ndata:')
        // Extract the JSON data chunk from the responseText
        const chunk = responseText.slice(lastIndex + 8)
        return chunk
}