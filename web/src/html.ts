export function html(tagName: string, attr: { [key: string]: {} } = {}) {
  const tag = document.createElement(tagName);
  for (const key in attr) {
    (tag as any)[key] = attr[key];
  }
  return tag;
}
