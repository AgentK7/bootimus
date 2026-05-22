import { defineConfig } from 'astro/config';
import node from '@astrojs/node';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';
import { visit } from 'unist-util-visit';

function rehypeStripDocMdLinks() {
  return (tree) => {
    visit(tree, 'element', (node) => {
      if (node.tagName !== 'a' || !node.properties) return;
      const href = node.properties.href;
      if (typeof href !== 'string') return;
      if (/^[a-z][a-z0-9+.-]*:/i.test(href) || href.startsWith('//') || href.startsWith('/') || href.startsWith('#')) return;
      node.properties.href = href.replace(/\.md(?=$|[#?])/, '');
    });
  };
}

export default defineConfig({
  site: 'https://bootimus.com',
  output: 'server',
  adapter: node({ mode: 'standalone' }),

  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'de', 'fr', 'es', 'ru', 'zh-CN'],
    routing: {
      prefixDefaultLocale: false,
      redirectToDefaultLocale: false,
    },
  },

  server: {
    host: '0.0.0.0',
    port: Number(process.env.PORT) || 3000,
  },

  compressHTML: true,

  build: {
    inlineStylesheets: 'auto',
  },

  markdown: {
    shikiConfig: {
      themes: { light: 'vitesse-light', dark: 'vitesse-dark' },
      defaultColor: false,
      wrap: true,
    },
    rehypePlugins: [
      rehypeStripDocMdLinks,
      rehypeSlug,
      [
        rehypeAutolinkHeadings,
        {
          behavior: 'append',
          properties: { className: ['heading-anchor'], 'aria-label': 'Permalink' },
          content: { type: 'text', value: '#' },
        },
      ],
    ],
  },
});
