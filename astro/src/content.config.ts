import { defineCollection, z } from 'astro:content';

const posts = defineCollection({
  type: 'content',
  schema: ({ image }) => z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.coerce.date(),
    draft: z.boolean().default(false),
    pinned: z.boolean().default(false),
    keywords: z.string().default(''),
    heroImage: image().optional(),
    heroAlt: z.string().default(''),
  }),
});

export const collections = { posts };
