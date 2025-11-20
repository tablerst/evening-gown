export function splitTextToSpans(element: HTMLElement, type: 'chars' | 'words' | 'lines' = 'chars') {
  const text = element.innerText
  element.innerHTML = ''

  if (type === 'chars') {
    const chars = text.split('')
    chars.forEach((char) => {
      const span = document.createElement('span')
      span.innerText = char
      span.style.display = 'inline-block'
      if (char === ' ') span.style.width = '0.3em' // preserve space width
      element.appendChild(span)
    })
    return Array.from(element.children)
  } else if (type === 'words') {
    const words = text.split(' ')
    words.forEach((word, index) => {
      const span = document.createElement('span')
      span.innerText = word
      span.style.display = 'inline-block'
      element.appendChild(span)
      if (index < words.length - 1) {
        const space = document.createTextNode(' ')
        element.appendChild(space)
      }
    })
    // Filter only element children (spans), ignore text nodes (spaces) for animation
    return Array.from(element.children)
  }
  
  return []
}

