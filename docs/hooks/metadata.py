"""Derive descriptions from existing documentation without changing its content."""

from copy import copy
from html.parser import HTMLParser

MAX_DESCRIPTION_LENGTH = 240
VOID_TAGS = {
    'area', 'base', 'br', 'col', 'embed', 'hr', 'img', 'input',
    'link', 'meta', 'param', 'source', 'track', 'wbr',
}
EXCLUDED_TAGS = {
    'aside', 'blockquote', 'details', 'figure', 'form', 'nav',
    'ol', 'pre', 'script', 'style', 'table', 'ul',
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
}


def normalize(text: str) -> str:
    return ' '.join(text.split())


class IntroductionParser(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.introduction = ''
        self.paragraph = None
        self.skipping = 0
        self.finished = False

    def handle_starttag(self, tag, attrs):
        if self.finished:
            return
        if tag in VOID_TAGS:
            if tag == 'br' and self.paragraph is not None and not self.skipping:
                self.paragraph.append(' ')
            return
        if self.skipping:
            self.skipping += 1
            return
        classes = (dict(attrs).get('class') or '').split()
        if tag in EXCLUDED_TAGS or 'admonition' in classes:
            self.skipping = 1
        elif tag == 'p':
            self.paragraph = []

    def handle_startendtag(self, tag, attrs):
        self.handle_starttag(tag, attrs)
        if tag not in VOID_TAGS:
            self.handle_endtag(tag)

    def handle_endtag(self, tag):
        if self.finished:
            return
        if self.skipping:
            self.skipping -= 1
        elif tag == 'p' and self.paragraph is not None:
            self.introduction = normalize(''.join(self.paragraph))
            self.paragraph = None
            self.finished = bool(self.introduction)

    def handle_data(self, data):
        if not self.finished and not self.skipping and self.paragraph is not None:
            self.paragraph.append(data)


def page_description(content: str, title: str, site_name: str) -> str:
    parser = IntroductionParser()
    parser.feed(content)
    parser.close()
    identity = normalize(title)
    if not parser.introduction:
        return f'{site_name}: {identity}.'
    description = f'{identity}. {parser.introduction}'
    if len(description) <= MAX_DESCRIPTION_LENGTH:
        return description
    end = description.rfind(' ', 0, MAX_DESCRIPTION_LENGTH - 2)
    return description[:end].rstrip() + '...' if end > 0 else description


def on_page_content(content: str, page, config, files) -> str:
    authored = page.meta.get('description')
    if authored is not None and not isinstance(authored, str):
        raise ValueError(f'{page.file.src_uri}: description must be a string.')
    description = normalize(authored) if authored is not None else ''
    if not description:
        if page.is_homepage and config.site_description:
            description = normalize(config.site_description)
        else:
            description = page_description(content, page.title, config.site_name)
    page.meta['description'] = description
    return content


def on_page_context(context, page, config, nav):
    from markupsafe import escape

    # MkDocs does not autoescape templates; keep source metadata plain text.
    rendered_page = copy(page)
    rendered_page.meta = {**page.meta, 'description': escape(page.meta['description'])}
    context['page'] = rendered_page
    return context
