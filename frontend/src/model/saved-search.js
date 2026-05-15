import RestModel from "model/rest";
import { $gettext } from "common/gettext";

// SavedSearch is a persisted photo search bookmark owned by the current user.
export class SavedSearch extends RestModel {
  getDefaults() {
    return {
      UID: "",
      Title: "",
      Query: "",
      Order: "",
      Reverse: false,
      Position: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  route() {
    const query = { q: this.Query };

    if (this.Order) {
      query.order = this.Order;
    }

    if (this.Reverse) {
      query.reverse = "true";
    }

    return { name: "browse", query: query };
  }

  static getCollectionResource() {
    return "saved-searches";
  }

  static getModelName() {
    return $gettext("Saved Search");
  }
}

export default SavedSearch;
