<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="560"
    class="p-dialog dialog-saved-search"
    color="background"
    @keydown.esc.exact="close"
    @keyup.enter.exact="confirm"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" class="form-saved-search" accept-charset="UTF-8" tabindex="-1" @submit.prevent="confirm">
      <v-card>
        <v-toolbar flat color="navigation" class="mb-4" density="comfortable">
          <v-toolbar-title>
            {{ $gettext(`Save Search`) }}
          </v-toolbar-title>
          <v-btn icon class="action-close" :aria-label="$gettext('Close')" @click.stop="close">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col cols="12">
              <v-text-field
                v-model="model.Title"
                hide-details
                autofocus
                :rules="[titleRule]"
                :label="$gettext('Title')"
                :disabled="disabled"
                class="input-title"
              ></v-text-field>
            </v-col>
            <v-col cols="12">
              <v-textarea
                v-model="model.Query"
                hide-details
                auto-grow
                rows="2"
                :rules="[queryRule]"
                :label="$gettext('Search Query')"
                :disabled="disabled"
                class="input-query"
              ></v-textarea>
            </v-col>
            <v-col cols="12" sm="8">
              <v-select
                v-model="model.Order"
                :label="$gettext('Sort Order')"
                :items="orderOptions"
                item-title="text"
                item-value="value"
                hide-details
                variant="solo-filled"
                :disabled="disabled"
                class="input-order"
              ></v-select>
            </v-col>
            <v-col cols="12" sm="4" class="d-flex align-center">
              <v-checkbox
                v-model="model.Reverse"
                :label="$gettext('Reverse')"
                :disabled="disabled || model.Order === 'random'"
                hide-details
                density="comfortable"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn variant="flat" color="highlight" class="action-confirm" :disabled="disabled" @click.stop="confirm">
            {{ $gettext(`Save`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import SavedSearch from "model/saved-search";
import $notify from "common/notify";

export default {
  name: "PSavedSearchDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    query: {
      type: String,
      default: "",
    },
    order: {
      type: String,
      default: "",
    },
    reverse: {
      type: [Boolean, String],
      default: false,
    },
  },
  emits: ["close", "confirm"],
  data() {
    return {
      disabled: !this.$config.allow("photos", "search"),
      model: new SavedSearch(),
      titleRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Name too long"),
      queryRule: (v) => !!v.trim() || this.$gettext("Required"),
      orderOptions: [
        { value: "", text: this.$gettext("Default") },
        { value: "newest", text: this.$gettext("Newest First") },
        { value: "oldest", text: this.$gettext("Oldest First") },
        { value: "random", text: this.$gettext("Random") },
      ],
    };
  },
  watch: {
    visible(show) {
      if (show) {
        this.model = new SavedSearch({
          Title: this.defaultTitle(),
          Query: this.query ? this.query.trim() : "",
          Order: this.order || "",
          Reverse: this.reverse === true || this.reverse === "true",
        });
      }
    },
  },
  methods: {
    defaultTitle() {
      const q = (this.query || "").trim();
      const match = q.match(/path:([^\s]+)/i);

      if (match && match[1]) {
        return match[1].replace(/^\//, "").replace(/\*$/, "");
      }

      if (q.length > 40) {
        return q.substring(0, 40) + "…";
      }

      return q || this.$gettext("Saved Search");
    },
    afterEnter() {
      this.$refs.dialog?.$el?.focus();
    },
    afterLeave() {
      this.$emit("close");
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.disabled) {
        return;
      }

      this.model
        .save()
        .then(() => {
          $notify.success(this.$gettext("Saved"));
          this.$emit("confirm", this.model);
          this.close();
        })
        .catch(() => {
          $notify.error(this.$gettext("Unable to save"));
        });
    },
  },
};
</script>
