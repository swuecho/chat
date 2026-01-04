const BLOCKED_TAGS = new Set(['script', 'iframe', 'object', 'embed'])
const BLOCKED_SVG_TAGS = new Set(['script', 'foreignobject'])

const stripUnsafeAttributes = (element: Element) => {
  for (const attr of Array.from(element.attributes)) {
    const name = attr.name.toLowerCase()
    const value = attr.value.trim().toLowerCase()

    if (name.startsWith('on')) {
      element.removeAttribute(attr.name)
      continue
    }

    if ((name === 'href' || name === 'src' || name === 'xlink:href') && value.startsWith('javascript:')) {
      element.removeAttribute(attr.name)
    }
  }
}

const sanitizeElementTree = (root: Element, blockedTags: Set<string>) => {
  const ownerDocument = root.ownerDocument || document
  const walker = ownerDocument.createTreeWalker(root, NodeFilter.SHOW_ELEMENT)
  const toRemove: Element[] = []

  let current = walker.currentNode as Element
  while (current) {
    const tagName = current.tagName.toLowerCase()
    if (blockedTags.has(tagName)) {
      toRemove.push(current)
    } else {
      stripUnsafeAttributes(current)
    }
    current = walker.nextNode() as Element
  }

  toRemove.forEach(node => node.remove())
}

export const sanitizeHtml = (input: string): string => {
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(input, 'text/html')
    sanitizeElementTree(doc.body, BLOCKED_TAGS)
    return doc.body.innerHTML
  } catch {
    return ''
  }
}

export const sanitizeSvg = (input: string): string => {
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(input, 'image/svg+xml')
    const root = doc.documentElement
    if (!root) return ''
    sanitizeElementTree(root, BLOCKED_SVG_TAGS)
    return new XMLSerializer().serializeToString(root)
  } catch {
    return ''
  }
}
