import {Link} from "react-router";

export default function NotBuilt() {
    return (
        <section className="card-cut border border-border bg-surface p-6">
            <h1 className="font-display text-xl font-bold">Not here yet</h1>
            <p className="mt-1 text-sm text-body">This page isn't built yet.</p>
            <Link to="/" className="mt-3 inline-block text-sm font-semibold text-brand">
                Back to Home
            </Link>
        </section>
    );
}
