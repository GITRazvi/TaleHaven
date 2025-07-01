import axios from "axios";

export default class Backend {
    private readonly BASEURL = "http://127.0.0.1:3001";


    constructor() {
    }

    async searchBooks(query: string): Promise<any> {
        const url = `${this.BASEURL}/search/books?query=${query}`;
        console.log("Requesting URL:", url);

        try {
            const response = await axios.get(url, { timeout: 10000 });
            return response.data;
        } catch (error: any) {
            console.error("Error while fetching books:", error.message || error);
            throw error;
        }
    }



}