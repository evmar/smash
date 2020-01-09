export function html(
  tagName: string,
  attr: { [key: string]: {} } = {},
  ...children: Node[]
) {
  const tag = document.createElement(tagName);
  for (const key in attr) {
    if (key === 'style') {
      const style = attr[key] as { [key: string]: {} };
      for (const key in style) {
        (tag.style as any)[key] = style[key];
      }
    } else {
      (tag as any)[key] = attr[key];
    }
  }
  for (const child of children) {
    tag.appendChild(child);
  }
  return tag;
}

export function htext(text: string): Node {
  return document.createTextNode(text);
}
