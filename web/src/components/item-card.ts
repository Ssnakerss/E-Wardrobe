/**
 * Item card component — displays a summary of an item.
 */

import { element } from '../dom/element.js';
import type { Item } from '../api/items.js';
import { formatPrice } from '../domain/item.js';

interface ItemCardProps {
  item: Item;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}

export function createItemCard(props: ItemCardProps): HTMLElement {
  const card = element({
    tag: 'article',
    className: 'item-card',
    attrs: { 'data-item-id': props.item.id },
  });

  // Thumbnail placeholder
  const hasPhotos = props.item.photos && props.item.photos.length > 0;
  const photo = hasPhotos ? props.item.photos[0] : null;
  card.appendChild(
    element({
      tag: 'div',
      className: 'item-card__image',
      children: photo
        ? [
            element({
              tag: 'img',
              attrs: {
                src: `/api/v1/photos/${photo.id}/content?variant=thumbnail`,
                alt: props.item.name,
                loading: 'lazy',
              },
            }),
          ]
        : [element({ tag: 'div', className: 'item-card__placeholder', children: ['No photo'] })],
    }),
  );

  // Content
  const content = element({
    tag: 'div',
    className: 'item-card__content',
    children: [
      element({
        tag: 'h3',
        className: 'item-card__title',
        children: [props.item.name],
      }),
      element({
        tag: 'p',
        className: 'item-card__brand',
        children: [props.item.brand || '—'],
      }),
      element({
        tag: 'p',
        className: 'item-card__price',
        children: [formatPrice(props.item.purchase_price, props.item.price_currency)],
      }),
      element({
        tag: 'p',
        className: 'item-card__meta',
        children: [
          props.item.season ? `${props.item.season.replace('_', ' ')} · ` : '',
          props.item.condition_rating ? `Condition: ${props.item.condition_rating}/5` : '',
        ].filter(Boolean) as (string | undefined)[],
      }),
      props.item.archived_at
        ? element({
            tag: 'span',
            className: 'item-card__badge item-card__badge--archived',
            children: ['Archived'],
          })
        : null,
    ].filter(Boolean),
  });

  card.appendChild(content);

  // Actions
  const actions = element({
    tag: 'div',
    className: 'item-card__actions',
    children: [
      props.onEdit
        ? element({
            tag: 'button',
            attrs: {
              type: 'button',
              className: 'btn btn--icon',
              'data-action': 'edit',
              'data-item-id': props.item.id,
            },
            children: ['Edit'],
          })
        : null,
      props.onDelete
        ? element({
            tag: 'button',
            attrs: {
              type: 'button',
              className: 'btn btn--icon btn--icon--danger',
              'data-action': 'delete',
              'data-item-id': props.item.id,
            },
            children: ['Delete'],
          })
        : null,
    ].filter(Boolean),
  });

  card.appendChild(actions);

  return card;
}
