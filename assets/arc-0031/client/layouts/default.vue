<template>
  <div class="w-screen h-screen">
    <NuxtLoadingIndicator :height="5" :duration="3000" />
    <slot />
  </div>
</template>
